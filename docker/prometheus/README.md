Prometheus Monitor
=====
Steps:
1. Run `docker-compose -f docker-prom-compose.yml up -d`.
2. Access to `http://localhost:3000`, login with `admin:admin`.
3. Add the data source of prometheus, typing url with `http://prom:9090`, then click the `Save&Test` button.
4. click the `Creat -> Import`, then click the `upload .json file` button to upload the `Go_Processes.json` file. 
