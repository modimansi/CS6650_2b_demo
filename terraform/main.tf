# Wire together four focused modules: network, ecr, logging, ecs.

module "network" {
  source         = "./modules/network"
  service_name   = var.service_name
  container_port = var.container_port
}

module "ecr" {
  source          = "./modules/ecr"
  repository_name = var.ecr_repository_name
}

module "logging" {
  source            = "./modules/logging"
  service_name      = var.service_name
  retention_in_days = var.log_retention_days
}

# Messaging infrastructure (SNS + SQS)
module "messaging" {
  source       = "./modules/messaging"
  service_name = var.service_name
}

# RDS MySQL Database
module "rds" {
  source                = "./modules/rds"
  service_name          = var.service_name
  vpc_id                = module.network.vpc_id
  private_subnet_ids    = module.network.private_subnet_ids
  ecs_security_group_id = module.network.ecs_security_group_id
  database_name         = var.database_name
  database_username     = var.database_username
  database_password     = var.database_password
}

# Reuse an existing IAM role for ECS tasks
data "aws_iam_role" "lab_role" {
  name = "LabRole"
}

# Simple ECS deployment (no ALB, no auto-scaling)
module "ecs" {
  source             = "./modules/ecs"
  service_name       = var.service_name
  image              = "${module.ecr.repository_url}:latest"
  container_port     = var.container_port
  subnet_ids         = module.network.public_subnet_ids  # Use public subnets for direct access
  security_group_ids = [module.network.ecs_security_group_id]
  execution_role_arn = data.aws_iam_role.lab_role.arn
  task_role_arn      = data.aws_iam_role.lab_role.arn
  log_group_name     = module.logging.log_group_name
  ecs_count          = 1  # Single task
  region             = var.aws_region
  # Explicitly set Fargate task size
  cpu                = "256"
  memory             = "512"
  # Pass SNS/SQS configuration as environment variables
  sns_topic_arn      = module.messaging.sns_topic_arn
  sqs_queue_url      = module.messaging.sqs_queue_url
  # Pass worker count for scaling payment processors
  worker_count       = var.worker_count
  # Pass database connection string
  database_url       = module.rds.connection_string
}


// Build & push the Go app image into ECR
resource "docker_image" "app" {
  # Use the URL from the ecr module, and tag it "latest"
  name = "${module.ecr.repository_url}:latest"

  build {
    # relative path from terraform/ → src/
    context = "../src"
    # Dockerfile defaults to "Dockerfile" in that context
  }
}

resource "docker_registry_image" "app" {
  # this will push :latest → ECR
  name = docker_image.app.name
}
