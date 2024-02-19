all:
	docker-compose up -d

build:
	cd src && go mod download
	cd ..

testAll:
	cd src && go test -v ./...
	cd ..

unitTest:
	cd src && go test -v main/tests/unit_test
	cd ..

apiTest:
	cd src && go test -v main/tests/api_test
	cd ..

loadTest:
	k6 run src/tests/load_test/script.js --log-output=none

clean:
	docker-compose down
