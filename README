# NATSY

An example project for learning how to use nats-server and async event patterns from Go services.

# Overview
Project contains two small Go services: Api-Producer & Consumer.
This repo is for learning purposes and to help us establish working norms when using this technology. 

### Api-Producer
Exposes API endpoints that cause async events to be created 
 - POST /products
    - Creates event on subject
    - `price.created`
 - GET /ping
    - Creates event on subjects
    - `ping.db.status`
    - `ping.api.status`
    - `ping.redis.status`

- API Request and Event models are clearly separated. 
- Higher level EventEnvelope captures lower level subject specific data


### Consumer
Simulates long running consumer for messages. Trying to show power and importance of subject naming through ping example
Subscribed to subjects:
- `ping.*.status` 
- `price.created`

Simulates a DB connection with a map

# Running
From main directory run 
> docker compose up --build