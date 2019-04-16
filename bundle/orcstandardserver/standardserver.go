package orcstandardserver

import (
	"github.com/steinarvk/orc"
	"github.com/steinarvk/orclib/lib/versioninfo"
	canonicalhost "github.com/steinarvk/orclib/module/orc-canonicalhost"
	orcephemeralkeys "github.com/steinarvk/orclib/module/orc-ephemeralkeys"
	httprouter "github.com/steinarvk/orclib/module/orc-httprouter"
	identity "github.com/steinarvk/orclib/module/orc-identity"
	identityapi "github.com/steinarvk/orclib/module/orc-identityapi"
	jsonapi "github.com/steinarvk/orclib/module/orc-jsonapi"
	logging "github.com/steinarvk/orclib/module/orc-logging"
	orcouterauth "github.com/steinarvk/orclib/module/orc-outerauth"
	orcpersistentkeys "github.com/steinarvk/orclib/module/orc-persistentkeys"
	persistentkeysapi "github.com/steinarvk/orclib/module/orc-persistentkeysapi"
	orcprometheus "github.com/steinarvk/orclib/module/orc-prometheus"
	publickeyregistry "github.com/steinarvk/orclib/module/orc-publickeyregistry"
	sectiontraceapi "github.com/steinarvk/orclib/module/orc-sectiontraceapi"
	server "github.com/steinarvk/orclib/module/orc-server"
)

var baseModules = []orc.Module{
	server.M,
	jsonapi.M,
	orcpersistentkeys.M,
	orcephemeralkeys.M,
	persistentkeysapi.M,
	identityapi.M,
	canonicalhost.M,
	httprouter.M,
	logging.M,
	orcouterauth.M,
	orcprometheus.M,
	publickeyregistry.M,
	sectiontraceapi.M,
}

var WithoutIdentity = &orc.ModuleBundle{
	Modules: baseModules,
}

func WithStandardIdentity() orc.Module {
	return &orc.ModuleBundle{
		Modules: append(baseModules, identity.MustNew(identity.Config{
			ProgramName: versioninfo.ProgramName,
			VersionInfo: versioninfo.MakeJSON(),
		})),
	}
}
