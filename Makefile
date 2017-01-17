
.PHONY: couchdb
couchdb:
	docker run -d -p 5984:5984 klaemo/couchdb:1.6.1

.PHONY: cov
cov:
	go test -v -coverprofile=coverage.out
	go tool cover -html=coverage.out
