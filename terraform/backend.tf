terraform {
  backend "s3" {
    bucket       = "terraform-states-ravi"
    key          = "event-receiver-service/terraform.tfstate"
    region       = "ap-south-1"
    use_lockfile = true
  }
}