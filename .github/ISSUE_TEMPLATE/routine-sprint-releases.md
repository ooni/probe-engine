---
name: Routine sprint releases
about: Weekly releases of probe-engine, etc.
title: ''
labels: effort/S, priority/medium
assignees: bassosimone

---

- [ ] Update dependencies
- [ ] Update internal/httpheader/useragent.go
- [ ] Update version.go
- [ ] Update resources/assets.go
- [ ] Run go generate ./...
- [ ] Tag a new version of ooni/probe-engine
- [ ] Create release at GitHub
- [ ] Update ooni/probe-engine mobile-staging branch
- [ ] Pin ooni/probe-cli to ooni/probe-engine
- [ ] Pin ooni/probe-android to latest mobile-staging
- [ ] Pin ooni/probe-ios to latest mobile-staging
