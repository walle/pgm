# pgm

Dead simple tool for running migrations for postgresql.

Supply a DSN to a database and a folder of .[up|down].sql files and run the up
command. The down command unmigrates one migration at the time.

This tool does not prevent you from shooting yourself in the foot, it only
executes the SQL you provide it with and saves a log to be able to undo work.

Motivation behind the tool is to have a slim docker image to be able to easily
run migrations using docker-compose in other projects. And also build new
images from the image that contains the migrations for running manually in
other environments i.e. prod.

## Installation

### Local and for development

`$ go get -u github.com/walle/pgm`

#### Building locally

`$ go install`, `$ go build` or `$ make pgm`

### Docker

#### Running locally

Run all migrations
```
$ docker run -v /my/migrations:/mymigrations --link mydb walle/pgm -dsn 'postgres://postgres:@mydb/postgres?sslmode=disable' -dir mymigrations up
```

Undo the latest migration
```
$ docker run -v /my/migrations:/mymigrations --link mydb walle/pgm -dsn 'postgres://postgres:@mydb/postgres?sslmode=disable' -dir mymigrations down
```

#### docker-compose

```yaml
version: '2'

services:
  db:
    restart: always
    image: postgres
    ports:
      - "5432:5432"

  migrations:
    image: walle/pgm
    command: -dsn 'postgres://postgres:@db/postgres?sslmode=disable' up
    volumes:
      - "./sql:/sql"
    links:
      - db
    depends_on:
      - db
```

See [docker-compose.yaml](docker-compose.yaml)

### Dependencies

The only dependency is [pq](https://github.com/lib/pq).

## Usage

`$ pgm -dsn postgres://postgres:@localhost/postgres?sslmode=disable -dir ./sql up`

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for more information.

## Versioning

We use [Semantic Versioning](http://semver.org/) for versioning. 
For the versions available, see the tags on this repository.

## Authors

* Fredrik Wallgren - https://github.com/walle

## License

This project is licensed under the MIT License - 
see the [LICENSE](LICENSE) file for details.
