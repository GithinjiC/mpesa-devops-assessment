terraform {
  backend "s3" {
    # Populated by `terraform init -backend-config=...`
    key     = "main/terraform.tfstate"
    encrypt = true
  }
}
