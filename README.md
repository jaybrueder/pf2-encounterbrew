# PF2 Encounterbrew

PF2 Encounterbrew is a self-hosted, mobile-browser friendly encounter tracker for the Pathfinder 2 TTRPG system by Paizo Inc.

## How do I run this

- Install [Docker](https://www.docker.com/)
- Download the [docker-compose.yml](./docker-compose.yml) file from this repository
- Open a terminal window
- Find the `docker-compose.yml` you just downloaded inside the terminal window (e.g. `cd ~/Downloads`)
- Run the following command:

```shell
docker compose up
```

You should now be able to access the application on [http://localhost:8080](http://localhost:8080)

If you have not changed the default username and password in the `docker-compose.yml` file, your can login with username `gamemaster` and password `changeme123`.

If you are done using the application, you can shut it down with typing `CTRL+C` in the terminal window.

All your encounters, parties and players will be stored inside a database and will still be there once your restart the application.

## Legal & copyright

"PF2 Encounterbrew" uses trademarks and/or copyrights owned by Paizo Inc., used under Paizo's Community Use Policy (paizo.com/communityuse). I am expressly prohibited from charging you to use or access this content. "PF2 Encounterbrew" is not published, endorsed, or specifically approved by Paizo. For more information about Paizo Inc. and Paizo products, visit paizo.com.

Creature data used in this encounter tracker, see `data` folder, is provided by the Pathfinder 2e FoundryVTT team under the Apache 2.0 license.

## FAQ

You got questions? I try to answer the most common ones here: [FAQ.md](./FAQ.md)

## Patch Notes

See [CHANGELOG.md](./CHANGELOG.md)

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md)
