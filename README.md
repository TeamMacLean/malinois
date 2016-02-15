# Malinois
> continuous deployment... thing

## Supported Continuous Integration Services

* Travis-ci.org

## Config

Config for Malinois
.malinois.yml
```
- travis: TeamMacLean/geefu.io
  dir: /opt/geefu.io
  action:
    - git pull
    - pm2 restart
- travis: TeamMacLean/datahog
  dir: /opt/datahog
  action:
    - git pull
    - pm2 restart
```

Config for Travis
.travis.yml
```
...
after_success:
    curl --data "repo=$TRAVIS_REPO_SLUG" http://127.0.0.1:8888
...
```