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

## Deploy on any machine (clean setup)

Prerequisites: Docker Desktop (running), Terraform, AWS CLI, an AWS account with permissions for ECR/ECS/VPC/CloudWatch Logs.

1) Configure AWS
```
aws configure            # or aws sso login --profile <profile>
$env:AWS_PROFILE = "<profile>"   # Windows PowerShell (optional)
```

2) Build and push (optional manual path)
```
# from repo root
docker build -t myapp:latest ./src

aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin 831956759147.dkr.ecr.us-east-1.amazonaws.com

docker tag myapp:latest 831956759147.dkr.ecr.us-east-1.amazonaws.com/ecr_service:latest
docker push 831956759147.dkr.ecr.us-east-1.amazonaws.com/ecr_service:latest
```

3) Provision or update infra (preferred)
```
cd terraform
terraform init
terraform apply -auto-approve
```

4) Find service public IP and test
```
# See above command in "Send Requests" to fetch the task ENI public IP
curl -i http://<PUBLIC-IP-ADDRESS>:8080/products/1
```

## Request/Response examples (all status codes)

Base URL: `http://<PUBLIC-IP-ADDRESS>:8080`

- GET `/products/1` → 200
```
curl -i http://<PUBLIC-IP-ADDRESS>:8080/products/1
```

- GET `/products/9999` → 404
```
curl -i http://<PUBLIC-IP-ADDRESS>:8080/products/9999
```

- POST `/products` → 201
```
curl -i -X POST http://<PUBLIC-IP-ADDRESS>:8080/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Widget 2","description":"Second product","price":12.49}'
```

- POST `/products` invalid body (missing name or negative price) → 400
```
curl -i -X POST http://<PUBLIC-IP-ADDRESS>:8080/products \
  -H "Content-Type: application/json" \
  -d '{"name":"","price":-1}'
```

- POST `/products/{id}/details` → 204
```
curl -i -X POST http://<PUBLIC-IP-ADDRESS>:8080/products/1/details \
  -H "Content-Type: application/json" \
  -d '{"name":"Widget Alpha","description":"First gen widget","price":12.49}'
```

- POST `/products/{id}/details` with negative price → 400
```
curl -i -X POST http://<PUBLIC-IP-ADDRESS>:8080/products/1/details \
  -H "Content-Type: application/json" \
  -d '{"price":-1}'
```

- POST `/products/{id}/details` where `{id}` does not exist → 404
```
curl -i -X POST http://<PUBLIC-IP-ADDRESS>:8080/products/9999/details \
  -H "Content-Type: application/json" \
  -d '{"name":"X"}'
```

Postman: Create an environment with `baseUrl = http://<PUBLIC-IP-ADDRESS>:8080` and add requests for each case above. You can export the collection (Collections → ... → Export) and share it with your team.

## Where things are

- Server code: `src/`
  - Product handlers: `src/product/*.go`
  - HTTP entrypoint: `src/main.go`
  - Dockerfile: `src/Dockerfile`
- Infrastructure (Terraform): `terraform/`
  - Main wiring: `terraform/main.tf`
  - Providers: `terraform/provider.tf`
  - Modules: `terraform/modules/*`
- Load testing: `testing/product_locustfile.py`

## Keeping large and sensitive files out of git

Already ignored:
```
**.tfvars
**.tfstate
**.tfstate.*
**/.terraform/
__pycache__/
*.pyc
*.log
```

Recommended: keep secrets in AWS Secrets Manager or local `.env` not checked in; use a separate, private `.tfvars` for environment-specific values.

## Clean Up
```
terraform destroy -auto-approve
```