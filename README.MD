## Instructions

### Prepare Credentials

Retrieve you temporary credentials from Learner's Lab.
Enter your configuration when prompted:
```
aws configure
```

And set your session token:
```
aws configure set aws_session_token <YOUR-TEMP-SESSION-TOKEBN>
```

### Apply Infrastructure
```
cd terraform
terraform init
terraform apply -auto-approve
```

### Send Requests
Make sure you are in terraform folder

Get public ip address
```
aws ec2 describe-network-interfaces \
--network-interface-ids $(
    aws ecs describe-tasks \
    --cluster $(terraform output -raw ecs_cluster_name) \
    --tasks $(
        aws ecs list-tasks \
        --cluster $(terraform output -raw ecs_cluster_name) \
        --service-name $(terraform output -raw ecs_service_name) \
        --query 'taskArns[0]' --output text
    ) \
    --query "tasks[0].attachments[0].details[?name=='networkInterfaceId'].value" \
    --output text
) \
--query 'NetworkInterfaces[0].Association.PublicIp' \
--output text
```

Send some requests:
```
curl http://<PUBLIC-IP-ADDRESS>:8080/albums
```

### Product API

Base URL: `http://<PUBLIC-IP-ADDRESS>:8080`

- Seeded data includes product `1`.
- New products are created via POST `/products` and get auto-incremented IDs.

Endpoints:

1) Create product

```bash
curl -i -X POST http://<PUBLIC-IP-ADDRESS>:8080/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Widget 2",
    "description": "Second product",
    "price": 12.49
  }'
```

Expected: `201 Created` with the new product JSON, e.g. `{ "id": 2, ... }`.

2) Get product by id

```bash
curl -i http://<PUBLIC-IP-ADDRESS>:8080/products/1
```

3) Update product details (for an existing id)

```bash
curl -i -X POST http://<PUBLIC-IP-ADDRESS>:8080/products/1/details \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Widget Alpha",
    "description": "First gen widget",
    "price": 12.49
  }'
```

Expected: `204 No Content`. Then verify with a GET to the same id.

Postman

- Create an environment variable `baseUrl` = `http://<PUBLIC-IP-ADDRESS>:8080`.
- Requests:
  - POST `{{baseUrl}}/products` with raw JSON body (application/json)
  - GET `{{baseUrl}}/products/1`
  - POST `{{baseUrl}}/products/1/details` with raw JSON body
- Tip: Use Postman’s "Code" button to copy cURL for each request. See [Postman](https://www.postman.com/).

## Clean Up
```
terraform destroy -auto-approve
```