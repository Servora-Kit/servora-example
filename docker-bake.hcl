variable "VERSION" {
  default = "dev"
}

target "_common" {
  context    = "."
  dockerfile = "Dockerfile"
  output = ["type=docker"]
  args = {
    VERSION = VERSION
  }
}

target "iam" {
  inherits = ["_common"]
  args = {
    SERVICE_NAME = "iam"
  }
  tags = [
    "servora/iam-service:${VERSION}",
    "servora/iam-service:latest",
  ]
}

target "sayhello" {
  inherits = ["_common"]
  args = {
    SERVICE_NAME = "sayhello"
  }
  tags = [
    "servora/sayhello-service:${VERSION}",
    "servora/sayhello-service:latest",
  ]
}

group "default" {
  targets = ["iam", "sayhello"]
}
