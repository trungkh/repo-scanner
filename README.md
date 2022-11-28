### Requirements
* Docker and Go

### Usage
Clone the repository with:
```bash
git clone gitlab.com/trungkh/repo-scanner
```

Copy the `env.example` file to a `.env` file.
```bash
$ cp .env.example .env
```

### Database Configurations
Update the postgres variables declared in the new `.env` to match your preference. 
There's a handy guide on the [Postgres' DockerHub](https://hub.docker.com/_/postgres).

### Auth Tokens
The use requires you to set various Git provider (Github / Gitlab / Bitbucket) authentication token in your environment.
You can do so by referring [Creating a personal github access token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token).

To persist the various Git provider tokens, you can change them among variables declared in `.env`.

### Skip Files
You can define paths to be excluded from scanning by defining them in a comma separated format in `.env` file.

`SKIP_EXT` defines the file extensions to be excluded
`SKIP_PATHS` defines the paths/files to be excluded if the path matches one of the patterns defined in the list
`SKIP_TEST_PATHS` defines any test directories/files that you would like to skip. It is being kept separately from `SKIP_PATHS` because sometimes it may be useful to scan the test files as well. You can toggle to scan test files by giving `-skip-tests=false` in the CLI.

### Boot Up
Build and start the containers with:
```bash
$ docker-compose up --build
```

### Run test cases
Running UTs:
```bash
$ go test -v ./...
```