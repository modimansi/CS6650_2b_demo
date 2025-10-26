output "ecs_cluster_name" {
  description = "Name of the created ECS cluster"
  value       = module.ecs.cluster_name
}

output "ecs_service_name" {
  description = "Name of the running ECS service"
  value       = module.ecs.service_name
}

output "vpc_id" {
  description = "ID of the custom VPC"
  value       = module.network.vpc_id
}

output "public_subnet_ids" {
  description = "IDs of public subnets for ECS tasks"
  value       = module.network.public_subnet_ids
}

output "private_subnet_ids" {
  description = "IDs of private subnets"
  value       = module.network.private_subnet_ids
}

output "sns_topic_arn" {
  description = "SNS Topic ARN for async orders"
  value       = module.messaging.sns_topic_arn
}

output "sqs_queue_url" {
  description = "SQS Queue URL for order processing"
  value       = module.messaging.sqs_queue_url
}

output "access_instructions" {
  description = "Instructions to access the ECS task"
  value       = <<-EOT
    To get the public IP of your ECS task, run:
    
    aws ecs describe-tasks \
      --cluster ${module.ecs.cluster_name} \
      --tasks $(aws ecs list-tasks --cluster ${module.ecs.cluster_name} --service-name ${module.ecs.service_name} --query 'taskArns[0]' --output text) \
      --query 'tasks[0].attachments[0].details[?name==`networkInterfaceId`].value' --output text | xargs -I {} aws ec2 describe-network-interfaces --network-interface-ids {} --query 'NetworkInterfaces[0].Association.PublicIp' --output text
    
    Then access your application at: http://<PUBLIC-IP>:8080
    
    Endpoints:
    - Sync:  POST http://<PUBLIC-IP>:8080/orders/sync  (blocks for 3s)
    - Async: POST http://<PUBLIC-IP>:8080/orders/async (returns immediately)
  EOT
}