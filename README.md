# discorddsb (WIP)

discorddsb aims to bring DSBmobile notifications to Discord via a Discord bot. 
At this moment this is very much WIP, see [Features/TODO](#featurestodo).

## Features/TODO
- [X] fetching/parsing substitute plans
  - [X] use [phlourishdev/DSBgo](https://pkg.go.dev/github.com/phlourishdev/DSBgo) to fetch substitute plan URLs
  - [X] fetch substitute plan HTMLs
  - [X] parse substitute plan HTMLs
- [ ] using a database for persistant storage
  - [X] prepare development environment with Docker
  - [X] initiate db connection
  - [ ] create tables, develop db model
  - [ ] add new entries to db
  - [ ] check if new entry already exists in db, if yes, skip
  - [ ] overwrite existing entries if data has changed
- [ ] add discord notifications via discord bot or webhook

## Building

1. Make sure you have Docker installed, if not, install Docker:
```shell
curl -fsSL https://get.docker.com | sh
```

2. Clone the repository
```shell
git clone https://github.com/phlourishdev/discorddsb.git
```

3. Build discorddsb via Docker Compose, you'll have to do this every time you edit the code
```shell
docker compose build
```

4. Start up the Docker containers
```shell
docker compose up
```

Don't forget to change `MYSQL_PASSWORD` and `MYSQL_ROOT_PASSWORD` in `compose.yml` as well as 
putting your DSB mobile credentials either in a `.env` or also in the `compose.yml`.