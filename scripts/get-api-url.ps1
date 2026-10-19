# Run after terraform apply. Read-only AWS queries; no infrastructure changes.
$ErrorActionPreference = 'Stop'
Push-Location (Join-Path $PSScriptRoot '../terraform')
try {
 function Get-TfOutput([string]$Name) {
  $value = & terraform output -raw $Name
  if ($LASTEXITCODE -ne 0) { throw "Cannot read Terraform output $Name. Run terraform apply first." }
  return ($value -join '').Trim()
 }
 function Invoke-AwsText([string[]]$Arguments) {
  $result = & aws @Arguments
  if ($LASTEXITCODE -ne 0) { throw 'AWS query failed. Check your lab credentials and region.' }
  return ($result -join '').Trim()
 }
 $region = Get-TfOutput 'aws_region'
 $cluster = Get-TfOutput 'ecs_cluster_name'
 $service = Get-TfOutput 'ecs_service_name'
 $port = Get-TfOutput 'container_port'
 $task = Invoke-AwsText @('ecs','list-tasks','--region',$region,'--cluster',$cluster,'--service-name',$service,'--desired-status','RUNNING','--query','taskArns[0]','--output','text')
 if (!$task -or $task -eq 'None') { throw 'No running ECS task. Check service events and CloudWatch logs.' }
 $eni = Invoke-AwsText @('ecs','describe-tasks','--region',$region,'--cluster',$cluster,'--tasks',$task,'--query',"tasks[0].attachments[?type=='ElasticNetworkInterface'].details[] | [?name=='networkInterfaceId'].value | [0]",'--output','text')
 if (!$eni -or $eni -eq 'None') { throw 'Task has no network interface yet. Wait, then retry.' }
 $ip = Invoke-AwsText @('ec2','describe-network-interfaces','--region',$region,'--network-interface-ids',$eni,'--query','NetworkInterfaces[0].Association.PublicIp','--output','text')
 if (!$ip -or $ip -eq 'None') { throw 'Task has no public IP. Check network configuration.' }
 "http://${ip}:${port}/v1"
} finally { Pop-Location }
