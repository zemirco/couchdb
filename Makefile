
couchdb:
	docker run -d -p 5984:5984 klaemo/couchdb:1.6.1

.PHONY: couchdb
