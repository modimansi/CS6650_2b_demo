# Lambda Module - AWS Learner Lab Compatible

This Lambda module is configured for **AWS Learner Lab / Academy** environments.

## Key Differences from Standard AWS

### IAM Role Restrictions

**AWS Learner Lab does NOT allow**:
- ❌ Creating IAM roles
- ❌ Creating IAM policies
- ❌ Attaching policies to roles

**Solution**: Use the pre-existing `LabRole` that has all necessary permissions.

### What This Module Does

1. **Uses LabRole**: References existing `LabRole` instead of creating new role
2. **Creates Lambda Function**: With Go runtime (provided.al2)
3. **SNS Subscription**: Triggers Lambda from SNS topic
4. **CloudWatch Logs**: Automatic logging
5. **Permissions**: Allows SNS to invoke Lambda

## Permissions Included in LabRole

LabRole already has permissions for:
- ✅ Lambda execution
- ✅ CloudWatch Logs
- ✅ SNS
- ✅ SQS
- ✅ ECS
- ✅ EC2
- ✅ S3
- ✅ And more...

## Deployment

```bash
# Build Lambda
cd lambda/payments_processor
./build.sh  # or .\build.ps1 on Windows

# Deploy
cd ../../terraform
terraform init
terraform apply
```

## Troubleshooting

### Error: "Not authorized to perform: iam:CreateRole"

**Fix**: Already applied! This module uses `LabRole` instead of creating roles.

### Error: "LabRole not found"

**Check**: Ensure you're in AWS Learner Lab environment
```bash
aws iam get-role --role-name LabRole
```

Should return role details with broad permissions.

## Resources Created

- ✅ Lambda Function
- ✅ SNS Topic Subscription
- ✅ Lambda Permission (for SNS)
- ✅ CloudWatch Log Group
- ❌ IAM Role (uses existing LabRole)
- ❌ IAM Policies (uses LabRole permissions)

## Configuration

All configuration is in `terraform/main.tf`:
- Memory: 512 MB (default)
- Timeout: 10 seconds
- Runtime: provided.al2 (Go)
- Trigger: SNS topic

## Cost

AWS Learner Lab has usage limits but no direct charges.

**Lambda Free Tier** (regular AWS):
- 1M requests/month free
- 400,000 GB-seconds free

After free tier: ~$0.20 per 1M requests

## Notes

This module is specifically designed for AWS Academy / Learner Lab constraints. For production AWS accounts, you may want to create dedicated IAM roles with least-privilege permissions.

