# pgm

Dead simple tool for running migrations for postgresql.

Supply a DSN to a database and a folder of .[up|down].sql files and run the up
command. The down command unmigrates one migration at the time.

## Installation

`$ go get -u github.com/walle/pgm`

### Building locally

`$ go build`

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
