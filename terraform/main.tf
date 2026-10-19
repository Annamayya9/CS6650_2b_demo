locals {
  # Changes to any image input produce a new tag and task definition.
  image_tag = substr(sha256(join("", [for file in ["product-service.go", "go.mod", "go.sum", "Dockerfile", ".dockerignore"] : filesha256("${path.module}/../src/${file}")])), 0, 16)
}

module "network" {
  source         = "./modules/network"
  service_name   = var.service_name
  container_port = var.container_port
  cidr_blocks    = var.allowed_cidrs
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
data "aws_iam_role" "execution" {
  name = var.execution_role_name
}
resource "docker_image" "app" {
  name = "${module.ecr.repository_url}:${local.image_tag}"
  build {
    context  = abspath("${path.module}/../src")
    platform = "linux/amd64"
  }
}
resource "docker_registry_image" "app" {
  name          = docker_image.app.name
  keep_remotely = true
}
module "ecs" {
  source             = "./modules/ecs"
  service_name       = var.service_name
  image              = "${module.ecr.repository_url}@${docker_registry_image.app.sha256_digest}"
  container_port     = var.container_port
  subnet_ids         = module.network.subnet_ids
  security_group_ids = [module.network.security_group_id]
  execution_role_arn = data.aws_iam_role.execution.arn
  task_role_arn      = null
  log_group_name     = module.logging.log_group_name
  ecs_count          = var.ecs_count
  region             = var.aws_region
  api_key            = var.api_key
  enable_test_errors = var.enable_test_errors
}
