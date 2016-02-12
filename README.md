# Malinois
> continuous deployment... thing


.malinois.yml example
```
- github: TeamMacLean/geefu.io
  travis: TeamMacLean/geefu.io
  dir: /opt/geefu.io
  action:
    - git pull
    - pm2 restart
- github: TeamMacLean/datahog
  travis: TeamMacLean/datahog
  dir: /opt/datahog
  action:
    - git pull
    - pm2 restart
```