There is account-svc microservice. It configured throw [config.yml](configs/config.yml). 
It persists [state](migrations/1.init.sql) to the PostgreSQL. It implements four [API](account.http) methods:
* create 
* credit
* debit
* getBalance

The objective of this task is to build a framework that achieves the following:
1. Launches the microservice for testing purposes within a Docker container.
2. Sets up the testing environment by:
   1. Configuring the 'config.yml' file.
   2. Preparing the database by running and applying migrations. 
3. Executing the tests.

The key requirement is to run the tests using a single command. There should be no need for manual steps or additional 
preparations beyond installing the framework itself. Please focus on the framework itself rather than on test cases.