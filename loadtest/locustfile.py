from locust import HttpUser, task, constant


class TodosApiUser(HttpUser):
    wait_time = constant(0)

    @task
    def list_todos(self):
        self.client.get('/todos')
