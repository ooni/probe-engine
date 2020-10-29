---
name: Routine sprint releases
about: Bi-weekly releases of probe-engine, etc.
title: ''
labels: effort/S, priority/medium
assignees: bassosimone

---

- [ ] engine: update dependencies
- [ ] engine: update internal/httpheader/useragent.go
- [ ] engine: update version.go
- [ ] engine: update resources/assets.go
- [ ] engine: update bundled certs (using `go generate ./...`)
- [ ] engine: make sure all workflows are green
- [ ] engine: tag a new version
- [ ] engine: update again version.go to be alpha
- [ ] engine: create release at GitHub
- [ ] engine: update mobile-staging branch to create oonimkall
- [ ] cli: pin to latest engine
- [ ] cli: update version/version.go
- [ ] cli: tag a new version
- [ ] cli: update version/version.go again to be alpha
- [ ] android: pin to latest oonimkall
- [ ] ios: pin to latest oonimkall
- [ ] desktop: pin to latest cli
- [ ] engine: create issue for next routine release
