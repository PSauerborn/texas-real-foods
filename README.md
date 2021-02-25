# Texas Real Foods: Data Collection and Aggregation Engine

Repository containing Backend codebase for the Texas Real Foods data collection and aggregation
engine. The codebase contains components to

1. RESTful API to manage the directory and businesses
2. RESTful API to retrieve business data and notifications
3. Data Collectors to collect business data from a variety of sources (Yelp, Google, Web Scraper etc)
4. NGINX routing components (Loadbalancers, routers, reverse proxy etc)

The majority of components are written in `Go`, with a few `Python` components/APIs, and the system follows
a micro-service architecture; all services and workers are run/deployed in `Docker` containers, and the entire
application can be started by running the docker compose file

```bash
$ docker-compose up --build --remove-orphans -d
```

from the root directory of the repository. note that this requires a running `Postgres` instance with
the relevant tables created

For more detailed documentation on the REST API exposed, visit https://trf.project-gateway.app/api/docs
to view the latest `Swagger` documentation for the current API endpoints. The following diagram illustrates
the architecture of the components

<br>

![Alt text](docs/architecture.png?raw=true "Title")