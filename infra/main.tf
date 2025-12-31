terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "ap-south-1"
}

# --- Variables ---
variable "instance_type" {
  description = "EC2 Instance Type"
  default     = "t3.micro"
  validation {
    condition     = contains(["t2.micro", "t3.micro"], var.instance_type)
    error_message = "FATAL ERROR: Instance type must be 't2.micro' or 't3.micro'."
  }
}

variable "volume_size" {
  description = "Root Volume Size in GB"
  default     = 30
  validation {
    condition     = var.volume_size <= 30
    error_message = "FATAL ERROR: Volume size cannot exceed 30GB."
  }
}

variable "db_backup_path" { default = "" }
variable "private_key_path" { default = "./cert-monitor-key.pem" }

# Secrets
variable "db_password" { sensitive = true }
variable "jwt_secret" { sensitive = true }
variable "smtp_user" {}
variable "smtp_pass" { sensitive = true }
variable "smtp_sender" {}

# --- 1. Security Group ---
resource "aws_security_group" "cert_sg" {
  name        = "cert-monitor-sg"
  description = "Allow Web and SSH"

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"] 
  }
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"] 
  }
  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"] 
  }
  ingress {
    from_port   = 81
    to_port     = 81
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"] 
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"] 

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-*"]
  }
}

# --- 2. Allocate Elastic IP (FIRST) ---
# We create this separately so we can know the IP address before the server starts
resource "aws_eip" "lb" {
  domain = "vpc"
}

# --- 3. EC2 Instance ---
resource "aws_instance" "app_server" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = var.instance_type
  key_name      = "cert-monitor-key" 

  vpc_security_group_ids = [aws_security_group.cert_sg.id]

  root_block_device {
    volume_size = var.volume_size
    volume_type = "gp3"
  }

  # Inject configuration
  user_data = templatefile("user_data.sh", {
    public_ip   = aws_eip.lb.public_ip
    db_password = var.db_password
    jwt_secret  = var.jwt_secret
    smtp_host   = "smtp-relay.brevo.com"
    smtp_port   = "587"
    smtp_user   = var.smtp_user
    smtp_pass   = var.smtp_pass
    smtp_sender = var.smtp_sender
  })

  tags = {
    Name = "CertMonitor-Auto"
  }
}

# --- 4. Associate IP to Instance (LAST) ---
# This glue connects the IP (Step 2) to the Instance (Step 3)
resource "aws_eip_association" "eip_assoc" {
  instance_id   = aws_instance.app_server.id
  allocation_id = aws_eip.lb.id
}

# --- 5. DB Restore Logic ---
resource "null_resource" "db_restore" {
  # Only run if a path is provided
  count = var.db_backup_path != "" ? 1 : 0

  # Connection info for the provisioners below
  connection {
    type        = "ssh"
    user        = "ubuntu"
    private_key = file(var.private_key_path)
    host        = aws_eip.lb.public_ip
    timeout     = "5m"
  }

  # Step 1: Upload the file from Laptop -> Server
  provisioner "file" {
    source      = var.db_backup_path
    destination = "/home/ubuntu/restore.sql"
  }

  # Step 2: Wait for Docker/DB to wake up, then restore
  provisioner "remote-exec" {
    inline = [
      "echo '⏳ [Restore] Waiting for Docker installation to finish...'",
      "while ! command -v docker &> /dev/null; do sleep 10; done",
      
      "echo '⏳ [Restore] Waiting for Containers to start...'",
      "until docker ps | grep cert_db; do sleep 5; done",
      
      "echo '⏳ [Restore] Waiting for Postgres to be ready...'",
      "sleep 10", 
      "until docker exec cert_db pg_isready -U postgres; do sleep 5; done",

      "echo '🚀 [Restore] Restoring Database from backup...'",
      "cat /home/ubuntu/restore.sql | docker exec -i cert_db psql -U postgres certdb",
      "echo '✅ [Restore] Database restoration complete!'"
    ]
  }

  depends_on = [aws_eip_association.eip_assoc]
}

output "server_ip" {
  value = aws_eip.lb.public_ip
}