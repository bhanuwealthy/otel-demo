SERVICE_NAME=otel
compose_ver=${docker-compose --version}
dcomp=docker-compose -f local-setup/local-dockercompose.yaml -p otel-project --profile
monitor_mode=On
app_port=9999
MONGO_ROOT=otelMongo

.PHONY: default full all check-composer hostname pip dirs run worker all-up otel db stopdb stopall killdb killworker


default: db run

full: pip all-up run
all: full

check-composer:
	docker-compose --version

hostname:
	echo 0.0.0.0 jarvis | sudo tee -a /etc/hosts

pip:
	sh install.sh

dirs:
	mkdir -p ${HOME}/projects/${MONGO_ROOT}     # This is for mongodb persistent volume
	mkdir -p ${HOME}/projects/tempo-data  # This is for tempo-traces persistent volume

# --------- Up commands ----------

run:
	APP_MONITORING=${monitor_mode} python run.py ${app_port} || echo '==========\n' && lsof -i:${app_port}


worker: check-composer db
	APP_MONITORING=${monitor_mode} celery --app ${SERVICE_NAME}.celery worker --loglevel=info --autoscale=20,3 -E &

all-up: dirs check-composer
	MONGO_ROOT=${MONGO_ROOT} ${dcomp} all up -d

otel: dirs check-composer
	${dcomp} otel up -d

db: dirs check-composer
	MONGO_ROOT=${MONGO_ROOT} ${dcomp} db up -d


# --------- kill commands ----------

stopdb:
	MONGO_ROOT=${MONGO_ROOT} ${dcomp} db stop

stopall:
	MONGO_ROOT=${MONGO_ROOT} ${dcomp} all stop

killdb: check-composer
	MONGO_ROOT=${MONGO_ROOT} ${dcomp} db down

killworker:
	pkill -f "${SERVICE_NAME}.celery worker"

killotel:
	${dcomp} otel down

killall:
	MONGO_ROOT=${MONGO_ROOT} ${dcomp} all down

# --------- Misc ----------


local-docs:
	docker run --rm -it -p 8000:8000 -v ${PWD}:/docs squidfunk/mkdocs-material

# Requires mapbox/some map-provider token to render the tiles
kepler:
	docker run --name kepler.gl --rm -it -p 8080:80 crazycapivara/kepler.gl

tools:
	brew install --cask mongodb-compass