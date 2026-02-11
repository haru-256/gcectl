terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~>7.14.1"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~>7.12.0"
    }
  }
}
