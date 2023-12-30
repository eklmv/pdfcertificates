db_image = postgres:16-alpine
db_test_container = postgres_pdfcert_test
db_test_port = 5432
db_test_user = postgres
db_test_password = secret
db_test_name = test_cert
db_test_url = "postgres://${db_test_user}:${db_test_password}@127.0.0.1:${db_test_port}/${db_test_name}?sslmode=disable"

.PHONY: docker.db.create
.ONESHELL:
docker.db.create:
	@ps=$$(docker ps -a -q -f "name=${db_test_container}")
	if [ -z $$ps ]; then
		docker create --name ${db_test_container} -p ${db_test_port}:5432  -e POSTGRES_PASSWORD=${db_test_password} -e POSTGRES_USER=${db_test_user} ${db_image} >/dev/null
		echo "${db_test_container} container created"
	else
		echo "${db_test_container} container already exits"
	fi

.PHONY: docker.db.rm
docker.db.rm: docker.db.down
	@docker rm ${db_test_container} >/dev/null
	echo "${db_test_container} container removed"

.PHONY: docker.db.up
docker.db.up: docker.db.create
	@docker start ${db_test_container} >/dev/null
	echo "${db_test_container} container started"

.PHONY: docker.db.down
docker.db.down:
	@docker stop ${db_test_container} >/dev/null
	echo "${db_test_container} container stopped"

.PHONY: docker.db.restart
docker.db.restart: docker.db.down docker.db.up

.PHONY: db.is_ready
.ONESHELL:
db.is_ready:
	@echo "waiting for ${db_test_container} container"
	docker exec ${db_test_container} sh -c "until pg_isready -q -h 127.0.0.1; do sleep 0.5; done"
	echo "${db_test_container} container is ready"

.PHONY: db.create
.ONESHELL:
db.create: docker.db.up db.is_ready
	@if docker exec ${db_test_container} sh -c "psql -U ${db_test_user} -lqt | cut -d \| -f 1 | grep -qw ${db_test_name}"; then
		echo "${db_test_name} database already exists"
	else
		docker exec ${db_test_container} createdb --username=${db_test_user} ${db_test_name}
		echo "${db_test_name} database created"
	fi

.PHONY: db.drop
db.drop: docker.db.up db.is_ready
	@docker exec ${db_test_container} dropdb --username=${db_test_user} ${db_test_name}
	echo "${db_test_name} database dropped"

.PHONY: psql
psql: docker.db.up db.is_ready db.create
	@docker exec -it ${db_test_container} psql -U ${db_test_user} ${db_test_name}

.PHONY: migrate.create
ONESHELL:
migrate.create: docker.db.up db.is_ready db.create
	@if [ -z $(n) ]; then
		echo "migrate.create require argument 'n' (name) to be set"
		exit 1
	else
		migrate create -ext sql -dir db/migrations -seq $(n)
	fi

.PHONY: migrate.up
migrate.up: docker.db.up db.is_ready db.create
	@migrate -path db/migrations -database ${db_test_url} up

.PHONY: migrate.down
migrate.down: docker.db.up db.is_ready db.create
	@migrate -path db/migrations -database ${db_test_url} down -all
