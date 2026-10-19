output "ecs_cluster_name" {
  description = "Name of the created ECS cluster"
  value       = module.ecs.cluster_name
}

output "ecs_service_name" {
  description = "Name of the running ECS service"
  value       = module.ecs.service_name
}
output "aws_region" {
  value = var.aws_region
}
output "container_port" {
  value = var.container_port
}
output "ecr_repository_url" {
  value = module.ecr.repository_url
}
output "log_group_name" {
  value = module.logging.log_group_name
}
output "image_tag" {
  value = local.image_tag
}
