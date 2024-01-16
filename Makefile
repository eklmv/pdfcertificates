ifneq (,$(wildcard ./test/test.env))
	include ./test/test.env
	export
endif

.SILENT:
.ONESHELL:

.PHONY: docker.db.create
docker.db.create:
	ps=$$(docker ps -a -q -f "name=${DB_TEST_CONTAINER}")
	if [ -z $$ps ]; then
		docker create --name ${DB_TEST_CONTAINER} -p ${DB_TEST_PORT}:5432  -e POSTGRES_PASSWORD=${DB_TEST_PASSWORD} -e POSTGRES_USER=${DB_TEST_USER} ${DB_IMAGE} >/dev/null \
		&& echo "${DB_TEST_CONTAINER} container created"
	else
		echo "${DB_TEST_CONTAINER} container already exists"
	fi

.PHONY: docker.db.rm
docker.db.rm: docker.db.down
	docker rm ${DB_TEST_CONTAINER} >/dev/null \
	&& echo "${DB_TEST_CONTAINER} container removed"

.PHONY: docker.db.up
docker.db.up: docker.db.create
	docker start ${DB_TEST_CONTAINER} >/dev/null \
	&& echo "${DB_TEST_CONTAINER} container started"

.PHONY: docker.db.down
docker.db.down:
	docker stop ${DB_TEST_CONTAINER} >/dev/null \
	&& echo "${DB_TEST_CONTAINER} container stopped"

.PHONY: docker.db.restart
docker.db.restart: docker.db.down docker.db.up

.PHONY: db.is_ready
db.is_ready:
	echo "waiting for ${DB_TEST_CONTAINER} container"
	docker exec ${DB_TEST_CONTAINER} sh -c "until pg_isready -q -h ${DB_TEST_IP}; do sleep 0.5; done" \
	&& echo "${DB_TEST_CONTAINER} container is ready"

.PHONY: db.create
db.create: docker.db.up db.is_ready
	if docker exec ${DB_TEST_CONTAINER} sh -c "psql -U ${DB_TEST_USER} -lqt | cut -d \| -f 1 | grep -qw ${DB_TEST_NAME}"; then
		echo "${DB_TEST_NAME} database already exists"
	else
		docker exec ${DB_TEST_CONTAINER} createdb --username=${DB_TEST_USER} ${DB_TEST_NAME} \
		&& echo "${DB_TEST_NAME} database created"
	fi

.PHONY: db.drop
db.drop: docker.db.up db.is_ready
	docker exec ${DB_TEST_CONTAINER} dropdb --username=${DB_TEST_USER} ${DB_TEST_NAME} \
	&& echo "${DB_TEST_NAME} database dropped"

.PHONY: psql
psql: docker.db.up db.is_ready db.create
	docker exec -it ${DB_TEST_CONTAINER} psql -U ${DB_TEST_USER} ${DB_TEST_NAME}

.PHONY: migrate.create
migrate.create: docker.db.up db.is_ready db.create
	if [ -z $(n) ]; then
		echo "migrate.create require argument 'n' (name) to be set"
		exit 1
	else
		migrate create -ext sql -dir db/migrations -seq $(n)
	fi

.PHONY: migrate.up
migrate.up: docker.db.up db.is_ready db.create
	migrate -path db/migrations -database ${DB_TEST_URL} up \
	&& echo "migrate up done"

.PHONY: migrate.down
migrate.down: docker.db.up db.is_ready db.create
	migrate -path db/migrations -database ${DB_TEST_URL} down -all \
	&& echo "migrate down done"

.PHONY: sqlc
sqlc:  migrate.up
	sqlc generate -f db/sqlc.yaml \
	&& echo "sqlc generate done"

.PHONY: test.cover
test.cover: docker.db.up db.is_ready
	mkdir -p out
	go test -coverprofile out/coverage.out -tags integration ./...

.PHONY: test.cover.html
test.cover.html:
	go tool cover -html out/coverage.out

.PHONY: test.unit.all
test.unit.all:
	go test ./... -timeout 30s

.PHONY: test.unit
test.unit:
	if [ -z $(run) ]; then
		echo "test.unit require argument 'run' to be set"
		exit 1
	else
		go test ./... -timeout 30s -run=$(run)
	fi


.PHONY: test.it.db.all
test.it.db.all: docker.db.up db.is_ready
	+ go test -count 1 -tags integration ./internal/db

.PHONY: test.it.db
test.it.db: docker.db.up db.is_ready
	if [ -z $(run) ]; then
		echo "migrate.it.db require argument 'run' to be set"
		exit 1
	else
		go test -count 1 -tags integration ./internal/db -run=$(run)
	fi

.PHONY: mock
mock:
	mockery --config test/mockery.yaml
