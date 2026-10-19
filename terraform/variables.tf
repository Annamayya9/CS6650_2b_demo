variable "aws_region" {
  type    = string
  default = "us-east-1"
}
variable "ecr_repository_name" {
  type    = string
  default = "assignment5-product-api"
}
variable "service_name" {
  type    = string
  default = "assignment5-product-api"
}
variable "container_port" {
  type    = number
  default = 8080
  validation {
    condition     = var.container_port >= 1 && var.container_port <= 65535 && floor(var.container_port) == var.container_port
    error_message = "container_port must be an integer from 1 to 65535."
  }
}
variable "ecs_count" {
  type    = number
  default = 1
  validation {
    condition     = var.ecs_count == 1
    error_message = "Use one task: this assignment stores products in memory, not in a shared database."
  }
}
variable "log_retention_days" {
  type    = number
  default = 7
}
variable "execution_role_name" {
  type        = string
  default     = "LabRole"
  description = "Existing role that ECS can assume to pull images and write logs."
}
variable "api_key" {
  type        = string
  sensitive   = true
  description = "API key supplied as TF_VAR_api_key. Stored in Terraform state and the ECS task definition."
  validation {
    condition     = length(var.api_key) >= 16
    error_message = "Choose an API key with at least 16 characters."
  }
}
variable "enable_test_errors" {
  type        = bool
  default     = false
  description = "Enable controlled 500 responses only when demonstrating the assignment."
}
variable "allowed_cidrs" {
  type        = list(string)
  default     = ["0.0.0.0/0"]
  description = "Who can reach the API; narrow to your public IPv4 /32 for a private demonstration."
}
variable "docker_host" {
  type        = string
  default     = null
  description = "Windows Docker Desktop: npipe:////./pipe/docker_engine. Linux/macOS: omit for provider default or set your engine socket."
}
