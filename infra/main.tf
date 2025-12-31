terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "ap-south-1" # Mumbai
}

# --- Variables ---

variable "instance_type" {
  description = "EC2 Instance Type"
  default     = "t3.micro"
  validation {
    condition     = contains(["t2.micro", "t3.micro"], var.instance_type)
    error_message = "FATAL ERROR: Instance type must be 't2.micro' or 't3.micro' to remain in AWS Free Tier."
  }
}

variable "volume_size" {
  description = "Root Volume Size in GB"
  default     = 30
  validation {
    condition     = var.volume_size <= 30
    error_message = "FATAL ERROR: Volume size cannot exceed 30GB (AWS Free Tier Limit)."
  }
}

# NEW: Path to the local .sql backup file (Optional)
# If left empty "", restore is skipped.
variable "db_backup_path" {
  description = "Local path to .sql backup file to restore (e.g., ./backups/backup.sql)"
  default     = ""
}

# NEW: Path to private key for file upload
variable "private_key_path" {
  description = "Path to the local private key .pem file"
  default     = "./cert-monitor-key.pem"
}

variable "db_password" { sensitive = true }
variable "jwt_secret" { sensitive = true }
variable "smtp_user" {}
variable "smtp_pass" { sensitive = true }
variable "smtp_sender" {}

# --- Resources ---

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

resource "aws_instance" "app_server" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = var.instance_type
  key_name      = "cert-monitor-key" 

  vpc_security_group_ids = [aws_security_group.cert_sg.id]

  root_block_device {
    volume_size = var.volume_size
    volume_type = "gp3"
  }

  user_data = templatefile("user_data.sh", {
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

resource "aws_eip" "lb" {
  instance = aws_instance.app_server.id
}

# --- 5. Conditional DB Restore ---
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

  # Ensure IP is attached before trying to SSH
  depends_on = [aws_eip.lb]
}

output "server_ip" {
  value = aws_eip.lb.public_ip
}