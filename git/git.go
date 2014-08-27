// git.go

// Package git provides custom domain access of github.com using go get and go import.
//
// adapted from https://github.com/niemeyer/gopkg/blob/master/main.go
// by David Skinner

package git

import (
	"appengine"
	"appengine/urlfetch"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var gogetTemplate = template.Must(template.New("").Parse(`
<html>
<head>
<meta name="go-import" content="{{.GopkgRoot}} git https://{{.GopkgRoot}}">
</head>
<body>
go get {{.GopkgPath}}
</body>
</html>
`))

type Repo struct {
	User         string
	Name         string
	SubPath      string
	OldFormat    bool
	MajorVersion Version
	AllVersions  VersionList
}

// GitHubRoot returns the repository root at GitHub, without a schema.
func (repo *Repo) GitHubRoot() string {
	return "github.com/davsk/" + repo.Name
}

// GopkgRoot returns the package root at gopkg.in, without a schema.
func (repo *Repo) GopkgRoot() string {
	return repo.GopkgVersionRoot(repo.MajorVersion)
}

// GopkgRoot returns the package path at gopkg.in, without a schema.
func (repo *Repo) GopkgPath() string {
	return repo.GopkgVersionRoot(repo.MajorVersion) + repo.SubPath
}

// GopkgVerisonRoot returns the package root in gopkg.in for the
// provided version, without a schema.
func (repo *Repo) GopkgVersionRoot(version Version) string {
	version.Minor = -1
	version.Patch = -1
	v := version.String()
	return "davsk.net/" + repo.Name + "." + v
}

// var patternOld = regexp.MustCompile(`^/(?:([a-z0-9][-a-z0-9]+)/)?((?:v0|v[1-9][0-9]*)(?:\.0|\.[1-9][0-9]*){0,2})/([a-zA-Z][-a-zA-Z0-9]*)(?:\.git)?((?:/[a-zA-Z][-a-zA-Z0-9]*)*)$`)
var patternNew = regexp.MustCompile(`^/(?:([a-zA-Z0-9][-a-zA-Z0-9]+)/)?([a-zA-Z][-.a-zA-Z0-9]*)\.((?:v0|v[1-9][0-9]*)(?:\.0|\.[1-9][0-9]*){0,2})(?:\.git)?((?:/[a-zA-Z][-a-zA-Z0-9]*)*)$`)

var httpClient *http.Client

// Handler does for davsk.net what gopkg.in does for everyone else.
func Handler(resp http.ResponseWriter, req *http.Request) {
	c := appengine.NewContext(req)
	httpClient = urlfetch.Client(c)

	m := patternNew.FindStringSubmatch(req.URL.Path)
	if m == nil {
		sendNotFound(resp, "Unsupported URL pattern; see the documentation at gopkg.in for details.")
		return
	}

	if strings.Contains(m[3], ".") {
		sendNotFound(resp, "Import paths take the major version only (.%s instead of .%s); see docs at gopkg.in for the reasoning.",
			m[3][:strings.Index(m[3], ".")], m[3])
		return
	}

	repo := &Repo{
		User:    m[1],
		Name:    m[2],
		SubPath: m[4],
	}

	var ok bool
	repo.MajorVersion, ok = parseVersion(m[3])
	if !ok {
		sendNotFound(resp, "Version %q improperly considered invalid; please warn the service maintainers.", m[3])
		return
	}

	var err error
	var refs []byte
	refs, repo.AllVersions, err = hackedRefs(repo)
	switch err {
	case nil:
		// all ok
	case ErrNoRepo:
		sendNotFound(resp, "GitHub repository not found at https://%s", repo.GitHubRoot())
		return
	case ErrNoVersion:
		v := repo.MajorVersion.String()
		sendNotFound(resp, `GitHub repository at https://%s has no branch or tag "%s", "%s.N" or "%s.N.M"`, repo.GitHubRoot(), v, v, v)
		return
	default:
		resp.WriteHeader(http.StatusBadGateway)
		resp.Write([]byte(fmt.Sprintf("Cannot obtain refs from GitHub: %v", err)))
		return
	}

	if repo.SubPath == "/git-upload-pack" {
		resp.Header().Set("Location", "https://"+repo.GitHubRoot()+"/git-upload-pack")
		resp.WriteHeader(http.StatusMovedPermanently)
		return
	}

	if repo.SubPath == "/info/refs" {
		resp.Header().Set("Content-Type", "application/x-git-upload-pack-advertisement")
		resp.Write(refs)
		return
	}

	resp.Header().Set("Content-Type", "text/html")
	if req.FormValue("go-get") == "1" {
		// execute simple template when this is a go-get request
		err = gogetTemplate.Execute(resp, repo)
		if err != nil {
			log.Printf("error executing go get template: %s\n", err)
		}
		return
	}

	renderPackagePage(resp, req, repo)
}

func sendNotFound(resp http.ResponseWriter, msg string, args ...interface{}) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	resp.WriteHeader(http.StatusNotFound)
	resp.Write([]byte(msg))
}

const refsSuffix = ".git/info/refs?service=git-upload-pack"

var ErrNoRepo = errors.New("repository not found in github")
var ErrNoVersion = errors.New("version reference not found in github")

func hackedRefs(repo *Repo) (data []byte, versions []Version, err error) {
	resp, err := httpClient.Get("https://" + repo.GitHubRoot() + refsSuffix)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot talk to GitHub: %v", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		// ok
	case 401, 404:
		return nil, nil, ErrNoRepo
	default:
		return nil, nil, fmt.Errorf("error from GitHub: %v", resp.Status)
	}

	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading from GitHub: %v", err)
	}

	var mrefi, mrefj int
	var vrefi, vrefj int
	var vrefv = InvalidVersion

	versions = make([]Version, 0)
	sdata := string(data)
	for i, j := 0, 0; i < len(data); i = j {
		size, err := strconv.ParseInt(sdata[i:i+4], 16, 32)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot parse refs line size: %s", string(data[i:i+4]))
		}
		if size == 0 {
			size = 4
		}
		j = i + int(size)
		if j > len(sdata) {
			return nil, nil, fmt.Errorf("incomplete refs data received from GitHub")
		}
		if sdata[0] == '#' {
			continue
		}

		hashi := i + 4
		hashj := strings.IndexByte(sdata[hashi:j], ' ')
		if hashj < 0 || hashj != 40 {
			continue
		}
		hashj += hashi

		namei := hashj + 1
		namej := strings.IndexAny(sdata[namei:j], "\n\x00")
		if namej < 0 {
			namej = j
		} else {
			namej += namei
		}

		name := sdata[namei:namej]

		if name == "refs/heads/master" {
			mrefi = hashi
			mrefj = hashj
		}

		if strings.HasPrefix(name, "refs/heads/v") || strings.HasPrefix(name, "refs/tags/v") {
			if strings.HasSuffix(name, "^{}") {
				// Annotated tag is peeled off and overrides the same version just parsed.
				name = name[:len(name)-3]
			}
			v, ok := parseVersion(name[strings.IndexByte(name, 'v'):])
			if ok && repo.MajorVersion.Contains(v) && (v == vrefv || !vrefv.IsValid() || vrefv.Less(v)) {
				vrefv = v
				vrefi = hashi
				vrefj = hashj
			}
			if ok {
				versions = append(versions, v)
			}
		}
	}

	// If there were absolutely no versions, and v0 was requested, accept the master as-is.
	if len(versions) == 0 && repo.MajorVersion == (Version{0, -1, -1}) {
		return data, nil, nil
	}

	if mrefi == 0 || vrefi == 0 {
		return nil, nil, ErrNoVersion
	}

	copy(data[mrefi:mrefj], data[vrefi:vrefj])
	return data, versions, nil
}
