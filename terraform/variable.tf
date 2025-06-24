variable "region" {
  default = "us-east-1"
}

variable "s3_bucket_name" {
  default = "your-name_portal_events"
}

variable "image" {
  default = "your-ecr-repo/event-receiver"
}

variable "subnets" {
  type = list(string)
}

variable "security_groups" {
  type = list(string)
}