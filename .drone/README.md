[Drone](https://www.drone.io/) is a CI tool integrated with github that we use for automated testing and building/publishing docker images.

# Table of Contents
1. [Release Preperation](#release-preperation)
2. [Local Development Drone Setup](#local-development-drone-setup)

# Release Preperation

When updating the drone.yml file the file must be re-signed using `make drone`. This is limited to Grafana employees for security reasons.

Drone environment variables will need to be setup beforehand.

```
export DRONE_SERVER=<url>
export DRONE_TOKEN=<token>
```

# Local Development Drone Setup

**IMPORTANT: IT MAY NOT BE SAFE TO RUN ALL PIPELINES LOCALLY WITHOUT MODIFICATION**

For validating your setup, the `Lint` or `Windows-Test` pipelines are safe to run depending on which architecture you are working on.

## **EASIER** (recommended)

Run the drone CLI locally following the drone documentation [here](https://docs.drone.io/cli/install/)

1. Install the CLI in a compatible container environment to the pipelines you want to test
1. Create a copy of .drone/drone.yml at the repo root called .drone.yml and copy the pipeline[s] you want to test in there
2. Example command from the repo root `drone exec --secret-file=secrets.txt --trusted --pipeline=Lint`

Pros
- Works with linux or windows containers depending on which environment it is installed on

Cons
- Doesn't exactly mirror running the pipeline on a drone server setup. For example some [environment](https://docs.drone.io/pipeline/environment/reference/) variables may not get set

## **HARDER** (not recommended)

Run the drone stack locally following the drone documentation [here](https://docs.drone.io/server/ha/developer-setup/)

1. Fork the repo
2. Start ngrok to get the public endpoint
3. Create and update the drone config files with the public endpoint
4. Configure an oath2 app in github with the public endpoint
5. Activate the repo in the drone server and configure it to point to .drone/drone.yml

Pros
- This will allow you to run the whole drone stack and test it end to end

Cons
- This doesn't include a windows drone agent so windows pipelines will not execute
- Every time you restart ngrok (getting a new public url) you will need to delete the containers, update the drone config, delete the webhook they made in github, update the oath2 app in github, recreat the containers, reactivate the repo in the drone server, etc