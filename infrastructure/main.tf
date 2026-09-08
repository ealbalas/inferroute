terraform {
  required_version = ">= 1.9"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  # Uncomment after creating the S3 bucket and DynamoDB table for state
  # backend "s3" {
  #   bucket         = "inferroute-terraform-state"
  #   key            = "prod/terraform.tfstate"
  #   region         = "us-west-2"
  #   dynamodb_table = "inferroute-terraform-locks"
  #   encrypt        = true
  # }
}

provider "aws" {
  region = var.aws_region
}

locals {
  name = "inferroute-${var.environment}"
  tags = {
    Project     = "inferroute"
    Environment = var.environment
    ManagedBy   = "terraform"
  }
}
