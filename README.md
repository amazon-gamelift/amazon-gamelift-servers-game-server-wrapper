- [Amazon GameLift Servers Game Server Wrapper](#amazon-gamelift-servers-game-server-wrapper)
- [Comparisons between Game Sever Wrapper and Full SDK Integration](#comparisons-between-game-sever-wrapper-and-full-sdk-integration)
- [Setup instructions](#setup-instructions)
  - [Prerequisites](#prerequisites)
  - [Build Steps](#build-steps)
- [Install and configuration instructions](#install-and-configuration-instructions)
  - [Anywhere Fleet](#anywhere-fleet)
  - [Managed EC2 Fleet](#managed-ec2-fleet)
  - [Managed Container Fleet](#managed-container-fleet)
- [Usage](#usage)
  - [Create Game Session](#create-game-session)
- [Metrics](#metrics)
- [Appendix](#appendix)
  - [Game Server Arguments](#game-server-arguments)
  - [Server SDK integration comparison against game server wrapper](#server-sdk-integration-comparison-against-game-server-wrapper)

# Amazon GameLift Servers Game Server Wrapper

This game server wrapper is designed for quick onboarding to [Amazon GameLift Servers](https://aws.amazon.com/gamelift/servers/) for any game server, ideal for rapid prototyping, fast iteration, and early testing. It allows you to run game server builds on Amazon GameLift Servers fleets without directly integrating your game server code with the server SDK for Amazon GameLift Servers. The wrapper supports the following types of Amazon GameLift Servers fleet deployments: Anywhere fleets, managed EC2 fleets, and managed container fleets.
Telemetry metrics collection is included for visibility into game server performance.

# Comparisons between Game Sever Wrapper and Full SDK Integration
The game server wrapper enables rapid onboarding to Amazon GameLift Servers, supports all three types of fleets and [game session queue feature](https://docs.aws.amazon.com/gamelift/latest/developerguide/queues-intro.html). However, it does not support all server SDK functionalities that you would get with full integration with the server SDK for Amazon GameLift Servers, for example defining customized server SDK callback functions and player session management. See details [here](https://console.aws.amazon.com/gamelift/hosting), and see [here](#server-sdk-apis-support-comparison-against-full-server-sdk-integration) for a comparison of Sever SDK API support between the two.

# Setup instructions
## Prerequisites
### Any Game Server Executable
A game sever executable is required.

### Go programming language, v1.18+
The game server wrapper is implemented in Go. Follow the Go installation guide for your platform: https://go.dev/doc/install

### AWS CLI
The AWS CLI will be used to create required Amazon GameLift Servers resources. See instructions on how to install the AWS CLI for each platform [here](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html).

### Make (required for Linux and Mac)

Make will be used to build the wrapper for Linux and Mac (a PowerShell script will be used for building for Windows).

```shell
# For Debian-based distributions like Ubuntu/Debian
sudo apt-get install make

# For RPM-based distributions like CentOS
sudo yum install make

# For Mac
brew install make
```

## Build Steps
### Mac and Linux
```shell
git clone git@github.com:amazon-gamelift/amazon-gamelift-servers-game-server-wrapper.git
cd amazon-gamelift-servers-game-server-wrapper
make
```

### Windows
```shell
git clone git@github.com:amazon-gamelift/amazon-gamelift-servers-game-server-wrapper.git
cd amazon-gamelift-servers-game-server-wrapper
powershell -file .\build.ps1
```

### Troubleshooting build issues
An error similar to `dial tcp: lookup proxy.golang.org: no such host` can be caused by a network or firewall issue when downloading dependencies. Run the following command before attempting to build again:
```shell
go env -w GOPROXY=direct
```

# Install and configuration instructions

After building the wrapper, in the `out` directory there will be artifacts built in three folders for each of the hosting options: `gamelift-servers-anywhere` for Anywhere fleets, `gamelift-servers-managed-ec2` for managed EC2 fleets and `gamelift-servers-managed-containers` for managed container fleets. See [Amazon GameLift Servers hosting options](https://docs.aws.amazon.com/gamelift/latest/developerguide/gamelift-intro-flavors.html#gamelift-intro-flavors-hosting) to learn more about the hosting solutions that Amazon GameLift Servers offers. You only need to pick one of the Anywhere / Managed EC2 / Manage Containers options at a time.

## Anywhere Fleet
### Configure AWS Profile
To use the Game Server Wrapper for Amazon GameLift Servers Anywhere fleets, it needs AWS credentials
to manage related Amazon GameLift Servers resources (for example [Amazon GameLift Servers computes](https://docs.aws.amazon.com/gamelift/latest/apireference/API_Compute.html)).

#### Option 1: Use SSO (Recommended)
##### 1. Get SSO credentials via IAM Identity Center
1. Navigate to **IAM Identity Center** in the AWS Console and click **Enable**
2. Verify Account ID and Region details, then click **Enable** again
3. Navigate to the **Users** dashboard and click **Add User**
4. Fill out User details, including Username, Email, First and Last Name
5. Check email to confirm password and complete user registration
6. In the **IAM Identity Center** console, navigate to the **Permission sets** dashboard and click **Create permission set**
7. Select the managed policy you would like to give your SSO user access to, click next, click next, then click create
6. In the **IAM Identity Center** console, navigate to the **AWS accounts** dashboard
7. Select the AWS Account that you would like to allow the SSO user access to, and click **Assign users or groups**
8. Find the User you created in the previous steps, select it, and click next
9. Find the Permission set you created in the previous steps, select it, and click next
10. Click **Submit** and wait for configuration to complete
11. Open terminal, run `aws configure sso` command, and provide the following details found in the **Settings summary** section of the **IAM Identity Center** dashboard:
```
$ aws configure sso
SSO session name (Recommended): <unique name for sso session>
SSO start URL [None]: <AWS access portal URL>
SSO region [None]: <region>
SSO registration scopes [None]: sso:account:access
```
12. Terminal will try to automatically open the verification link in your browser. If this does not happen automatically, copy the link in the terminal into your browser
13. Sign in to your SSO user account using the Username and Password created during registration
14. Navigate back to the terminal and provide the following details:
```
CLI default client Region [us-west-2]: <region>
CLI default output format [None]: json
CLI profile name [Role-Account]: <profile name>
```
Your SSO credentials will be stored in your `~/.aws` directory. You can provide the SSO profile name provided in the previous step in the `profile` field of the wrapper `config.yaml` file to use this SSO account for Anywhere resource configuration.
##### 2. Get SSO credentials via other provider
If you are using another SSO credential provider, update the wrapper `config.yaml` `provider` field to `sso-file`, and provide the path to your SSO credential file in the `profile` field.

#### Option 2: Use shared credentials file

##### 1. Generate credentials
1. Create your IAM user by following the [Creating IAM users (console)](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_create.html#id_users_create_console) procedure in the IAM User Guide.
    1. For Permission options, choose Attach policies directly for how you want to assign permissions to this user.
    2. To provide the wrapper application with full access to Amazon GameLift Servers, select Create Policy and use the following policy:
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
               "gamelift:ListCompute",
               "gamelift:RegisterCompute",
               "gamelift:DeregisterCompute",
               "gamelift:GetComputeAuthToken"
            ],
            "Resource": "*"
        }
    ]
}
```

2. Get your access keys
    1. Sign in to the AWS Management Console and open the IAM console at https://console.aws.amazon.com/iam/
    2. In the navigation pane of the IAM console, select Users and then select the username of the user that you created previously.
    3. On the user's page, select the Security credentials page. Then, under Access keys, select Create access key.
    4. For Step 1 of Create access key, choose Other.
    5. For Step 2 of Create access key, enter an optional tag and select Next.
    6. For Step 3 of Create access key, select Download .csv file to save a .csv file with your IAM user's access key and secret access key. You need this information for later.
    7. Select Done.

##### 2. Configure profile

The `aws configure` command is the fastest way to set up your profile. This configure wizard prompts you for each piece of information you need to get started.

The following example configures a default profile using sample values. Replace them with your own values as described in the following sections.

```
$ aws configure --profile user001
AWS Access Key ID [None]: AAABBBCCC1111EXAMPLE
AWS Secret Access Key [None]: AAAAAAAAAAAAA/BBBBBBB/CCCCCCCCEXAMPLEKEY
Default region name [None]: us-west-2
Default output format [None]: json
```

### Prepare the game directory

- Find the build artifact for Anywhere. The build will be written to an out directory named `gamelift-servers-anywhere` for example: `out\linux\amd64\gamelift-servers-anywhere`.
- Copy the game server into the `gamelift-servers-anywhere` folder
- You can copy all content of the folder to any compatible machine or any other directory as needed.

An example directory structure will look like this:

```
gamelift-servers-anywhere
│-- config.yaml
│-- amazon-gamelift-servers-game-server-wrapper
│-- MyGame
│   │-- my-server-executable
│   │-- my-game-settings
│   │ ......
```

NOTE: For Linux machines, you might need to configure the permissions using `chown -R username <DIRECTORY>` and then `chmod +x <PATH_TO_GAME_EXECUTABLE>`.

For example if you moved the wrapper and game executable to `/local/game`, use
```bash
chown -R username /local/game
chmod +x /local/game/MyGame/my-server-executable
```

### Create Anywhere resources
Follow the `Create a custom location` step and `Create an Anywhere fleet` step in [Create an Amazon GameLift Servers Anywhere fleet](https://docs.aws.amazon.com/gamelift/latest/developerguide/fleets-creating-anywhere.html) to create a custom location and Anywhere fleet. Get their resources ARNs for wrapper configurations in the next step.

Optionally, additionally follow the `Add a compute to the fleet` step to register a [compute](https://docs.aws.amazon.com/gamelift/latest/apireference/API_Compute.html) resource. Get the `ComputeName` and `GameLiftServiceSdkEndpoint` for wrapper configurations in the next step.

### Configure Wrapper

Configuration for the wrapper is the config.yaml file. Edit this file following the instruction:

#### Example Configuration
```yaml
log-config:
  wrapper-log-level: debug                                      # The verbosity of the log messages emitted by the wrapper. Valid options are: debug, info, warn, error.

anywhere:
  provider: aws-profile                                         # The provider type for your AWS credentials (either [aws-profile] or [sso-file])
                                                                # If using IAM Identity Center for SSO Credentials, this should still be aws-profile
                                                                # If using other SSO provider, this should be sso-file

  profile: user001                                              # If using aws-profile provider, this will be the name of the aws profile configured in .aws/config
                                                                # If using IAM Identity Center for SSO Credentials, this should be the profile name created as part of CLI configuration
                                                                # If using other SSO provider, this should be the path to your SSO credentials file

  location-arn: arn:aws:gamelift:us-west-2-your-location-arn    # The AWS Arn of the location
  fleet-arn: arn:aws:gamelift:us-west-2-your-fleet-arn          # The AWS Arn of the Anywhere fleet
  ipv4: 127.0.0.1                                               # The IP address of the machine
  compute-name: DevLaptop                                       # (Optional) The name of an already registered compute
  service-sdk-endpoint: wss://us-west-2.api.amazongamelift.com  # (Optional) The ServiceSdkEndpoint on an already registered compute to be used for communicating with Amazon GameLift Servers

ports:
  gamePort: 37016

game-server-details:
  executable-file-path: ./MyGame/my-server-executable           # Entry point to execute the game server
  game-server-args:                                             # (Optional) Argument key value pairs that are passed to the game server entry point
    - arg: "--port"
      val: "{{.GamePort}}"
      pos: 0
```

- The `aws-profile` will be the profile name that you configured in [Configure AWS Profile](#configure-aws-profile).
- Use the resources ARNs generated in [Create Anywhere resources](#create-anywhere-resources) for `location-arn` and `fleet-arn`.
- (Optional) Use the compute resource output generated in [Create Anywhere resources](#create-anywhere-resources) for `compute-name` and `service-sdk-endpoint` to prevent the wrapper from registering a new Compute resource using your machine's `hostname`.
- Provide the path of your game server executable in `executable-file-path`. Using the above config as an example the wrapper would expect the game server to be on disk at `./gameserver.sh`
- `game-server-args` defines arguments that will be passed to the game server executable. See [Game Server Arguments](#game-server-arguments) for details.

### Launch the Wrapper
You can launch the wrapper by executing `amazon-gamelift-servers-game-server-wrapper`, for example executing in the terminal:
```bash
./amazon-gamelift-servers-game-server-wrapper
```

The wrapper will register your machine to Amazon GameLift Servers and establish connections with Amazon GameLift Servers. When the wrapper receives a signal to create a game session, it will start the server executable automatically.

You can launch multiple instances of the wrapper to have one machine host multiple concurrent game sessions by providing different ports when executing `amazon-gamelift-servers-game-server-wrapper`, for example executing in the terminal:
```bash
./amazon-gamelift-servers-game-server-wrapper --port 37000
```

If you have set a `compute-name` and `service-sdk-endpoint` in config.yaml, you can optionally additionally provide an already requested authorization token when executing `amazon-gamelift-servers-game-server-wrapper` to prevent the wrapper from requesting a new authorization token, for example executing in the terminal:
```bash
./amazon-gamelift-servers-game-server-wrapper --auth-token 00000000-1111-2222-3333-444444444444
```

## Managed EC2 Fleet
### Prepare the game directory

- Find the build artifact for Managed EC2. The build will be written to an out directory named `gamelift-servers-managed-ec2` for example: `out\linux\amd64\gamelift-servers-managed-ec2`.
- Copy the game server into the `gamelift-servers-managed-ec2` folder

An example directory structure will look like this:

```
gamelift-servers-managed-ec2
│-- config.yaml
│-- amazon-gamelift-servers-game-server-wrapper
│-- MyGame
│   │-- my-server-executable
│   │-- my-game-settings
│   │ ......
```

### Configure Wrapper

Configuration for the wrapper is the config.yaml file. Edit this file following the instruction:

#### Example Configuration

```yaml
log-config:
  wrapper-log-level: debug                                      # The verbosity of the log messages emitted by the wrapper. Valid options are: debug, info, warn, error.
  game-server-logs-dir: ./game-server-logs                      # (Optional) path where game server logs are written, will be uploaded to gamelift.

ports:
  gamePort: 37016

game-server-details:
  executable-file-path: ./MyGame/my-server-executable           # Entry point to execute the game server
  game-server-args:                                             # (Optional) Argument key value pairs that are passed to the game server entry point
    - arg: "--port"
      val: "{{.GamePort}}"
      pos: 0
```

- You can define the path of game server logs in `game-server-logs-dir`. The content will be uploaded to Amazon GameLift Servers, and will be available to download using [GetGameSessionLogUrl API](https://docs.aws.amazon.com/gamelift/latest/apireference/API_GetGameSessionLogUrl.html).
- Provide the path of your game server executable in `executable-file-path`. Using the above config as an example the wrapper would expect the game server to be on disk at `./gameserver.sh`
- `game-server-args` defines arguments that will be passed to the game server executable. See [Game Server Arguments](#game-server-arguments) for details

### Upload the build to Managed EC2 Fleet

Using [upload-build](https://docs.aws.amazon.com/cli/latest/reference/gamelift/upload-build.html) to upload the build to Amazon GameLift Servers

Example:
```bash
aws gamelift upload-build \
    --name gamelift-test-2025-03-11-1 \
    --build-version gamelift-test-2025-03-11-1 \
    --build-root out/linux/amd64/gamelift-servers-managed-ec2 \
    --operating-system AMAZON_LINUX_2023 \
    --server-sdk-version 5.4.0 \
    --region us-west-2
```
- For Windows builds, use `--operating-system WINDOWS_2016`

After creating the build, record the **build id** from the API response. We will use it for fleet creation.

### Create a Managed EC2 Fleet

Create a Managed EC2 fleet using the [CreateFleet API](https://docs.aws.amazon.com/gamelift/latest/apireference/API_CreateFleet.html).

To create a fleet that has two or more concurrent game server processes each with a separate port, set the `--port` field in the RuntimeConfiguration
ServerProcess Parameters as well as setting the EC2 Inbound Permissions range to include those ports.

Mac and Linux example:
```bash
aws gamelift create-fleet \
    --name 'gamelift-test-2025-03-11-1' \
    --description 'gamelift-test-2025-03-11-1' \
    --build-id build-ff1eb06c-74ab-4c42-aa99-example \
    --ec2-instance-type c5.large \
    --ec2-inbound-permissions 'FromPort=37000,ToPort=37020,IpRange=0.0.0.0/0,Protocol=UDP' \
                              'FromPort=37000,ToPort=37020,IpRange=0.0.0.0/0,Protocol=TCP' \
    --compute-type EC2 \
    --region us-west-2 \
    --runtime-configuration '{
        "ServerProcesses": [
            {
                "LaunchPath": "/local/game/amazon-gamelift-servers-game-server-wrapper",
                "ConcurrentExecutions": 1,
                "Parameters": "--port 37000"
            },
            {
                "LaunchPath": "/local/game/amazon-gamelift-servers-game-server-wrapper",
                "ConcurrentExecutions": 1,
                "Parameters": "--port 37001"
            }
        ]
    }'
```

Windows PowerShell example:
```powershell
aws gamelift create-fleet `
    --name 'gamelift-test-2025-03-11-1' `
    --description 'gamelift-test-2025-03-11-1' `
    --build-id build-ff1eb06c-74ab-4c42-aa99-example `
    --ec2-instance-type c5.large `
    --ec2-inbound-permissions 'FromPort=37000,ToPort=37020,IpRange=0.0.0.0/0,Protocol=UDP' `
                              'FromPort=37000,ToPort=37020,IpRange=0.0.0.0/0,Protocol=TCP' `
    --compute-type EC2 `
    --region us-west-2 `
    --runtime-configuration 'ServerProcesses=[{LaunchPath="C:\game\amazon-gamelift-servers-game-server-wrapper.exe",ConcurrentExecutions=1,Parameters="--port 37000"},{LaunchPath="C:\game\amazon-gamelift-servers-game-server-wrapper.exe",ConcurrentExecutions=1,Parameters="--port 37001"}]'
```

- `build-id` should match the build that was created in the last step.

After creating the fleet, record the **fleet id** from the API response. We will use it for game session creations.

Wait for it to be ACTIVE before proceeding. You can check its status using [AWS console](https://console.aws.amazon.com/gamelift/ec2-fleets) or [DescribeFleetAttributes API](https://docs.aws.amazon.com/gamelift/latest/apireference/API_DescribeFleetAttributes.html)

## Managed Container Fleet

**Amazon GameLift Servers only supports Linux containers. Build the wrapper and the game server for Linux before proceeding.**

Docker is required for deploying container fleets. See https://docs.docker.com/desktop/ for Docker installation guide.

### Prepare the game directory

- Find the build artifact for Containers. The build will be written to an out directory named `gamelift-servers-managed-containers` for example: `out\linux\amd64\gamelift-servers-managed-containers`.
- Copy the game server into the `gamelift-servers-managed-containers` folder

An example directory structure will look like this:

```
gamelift-servers-managed-containers
│-- config.yaml
│-- Dockerfile
│-- amazon-gamelift-servers-game-server-wrapper
│-- MyGame
│   │-- my-server-executable
│   │-- my-game-settings
│   │ ......
```

### Configure Wrapper

Configuration for the wrapper is the config.yaml file. Edit this file following the instruction:

#### Example Configuration

```yaml
log-config:
  wrapper-log-level: debug                                      # The verbosity of the log messages emitted by the wrapper. Valid options are: debug, info, warn, error.

ports:
  gamePort: 37016

game-server-details:
  executable-file-path: ./MyGame/my-server-executable           # Entry point to execute the game server
  game-server-args:                                             # (Optional) Argument key value pairs that are passed to the game server entry point
    - arg: "--port"
      val: "{{.ContainerPort}}"                                 # This is the container-internal port that your game server listens on. The value comes from the 'ports: gamePort' setting.
      pos: 0
```

- Provide the path of your game server executable in `executable-file-path`. Using the above config as an example the wrapper would expect the game server to be on disk at `./gameserver.sh`
- `game-server-args` defines arguments that will be passed to the game server executable. See [Game Server Arguments](#game-server-arguments) for details.

### Build Image
Build the image using `docker build` for the gamelift-servers-managed-containers directory. For example
```bash
docker build --platform=linux/amd64 -t gamelift-sdk-wrapper-sample:latest out/linux/amd64/gamelift-servers-managed-containers --progress=plain
```
### Push Image to ECR
In this step we will use AWS CLI to create an [Amazon ECR private repository](https://docs.aws.amazon.com/AmazonECR/latest/userguide/Repositories.html), and push the image we built from the last step to the repository. You must modify `AWS_ACCOUNT_ID` to be your account ID.

Mac and Linux:
```bash
export AWS_ACCOUNT_ID=123456789012
aws ecr create-repository --repository-name gamelift-sdk-wrapper-sample --region us-west-2
aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin ${AWS_ACCOUNT_ID}.dkr.ecr.us-west-2.amazonaws.com
docker tag gamelift-sdk-wrapper-sample:latest ${AWS_ACCOUNT_ID}.dkr.ecr.us-west-2.amazonaws.com/gamelift-sdk-wrapper-sample:latest
docker push ${AWS_ACCOUNT_ID}.dkr.ecr.us-west-2.amazonaws.com/gamelift-sdk-wrapper-sample:latest
```

Windows PowerShell:
```powershell
$AWS_ACCOUNT_ID = "123456789012"
aws ecr create-repository --repository-name gamelift-sdk-wrapper-sample --region us-west-2
aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin "$AWS_ACCOUNT_ID.dkr.ecr.us-west-2.amazonaws.com"
docker tag gamelift-sdk-wrapper-sample:latest "$AWS_ACCOUNT_ID.dkr.ecr.us-west-2.amazonaws.com/gamelift-sdk-wrapper-sample:latest"
docker push "$AWS_ACCOUNT_ID.dkr.ecr.us-west-2.amazonaws.com/gamelift-sdk-wrapper-sample:latest"
```

### Create ContainerGroupDefinition
Now we create an Amazon GameLift Servers container group definition using the [CreateContainerGroupDefinition API](https://docs.aws.amazon.com/gamelift/latest/apireference/API_CreateContainerGroupDefinition.html).
```
aws gamelift create-container-group-definition \
--region us-west-2 \
--name "gamelift-sdk-wrapper-sample" \
--operating-system AMAZON_LINUX_2023 \
--total-memory-limit-mebibytes 1024 \
--total-vcpu-limit 1 \
--game-server-container-definition "{\"ContainerName\": \"GameServer\", \"ImageUri\": \"${AWS_ACCOUNT_ID}.dkr.ecr.us-west-2.amazonaws.com/gamelift-sdk-wrapper-sample:latest\", \"PortConfiguration\": {\"ContainerPortRanges\": [{\"FromPort\": 37016, \"ToPort\": 37020, \"Protocol\": \"TCP\"}, {\"FromPort\": 37016, \"ToPort\": 37020, \"Protocol\": \"UDP\"}]}, \"ServerSdkVersion\": \"5.4.0\"}"
```
Wait for it to be READY before proceeding. You can check its status using [AWS console](https://console.aws.amazon.com/gamelift/container-groups) or [DescribeContainerGroupDefinition API](https://docs.aws.amazon.com/gamelift/latest/apireference/API_DescribeContainerGroupDefinition.html)
### Create Container Fleet Role
An IAM service role is required for Container fleets for Amazon GameLift Servers to manage your resources. You can create one using the following commands. See [Create an IAM role for Amazon GameLift Servers managed containers
](https://docs.aws.amazon.com/gamelift/latest/developerguide/setting-up-role.html#setting-up-role-containers) for details.
```
aws iam create-role --role-name GameLiftContainerFleet --assume-role-policy-document '{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":"gamelift.amazonaws.com"},"Action":"sts:AssumeRole"}]}'
aws iam attach-role-policy --role-name GameLiftContainerFleet --policy-arn arn:aws:iam::aws:policy/GameLiftContainerFleetPolicy
```

### Create Container Fleet
Create a container fleet using [CreateContainerFleet API](https://docs.aws.amazon.com/gamelift/latest/apireference/API_CreateContainerFleet.html):
```bash
aws gamelift create-container-fleet \
--region us-west-2 \
--description "Test container fleet" \
--instance-type c5.large \
--fleet-role-arn "arn:aws:iam::${AWS_ACCOUNT_ID}:role/GameLiftContainerFleet" \
--game-server-container-group-definition-name "gamelift-sdk-wrapper-sample"
```
After creating the fleet, record the **fleet id** from the API response. We will use it for game session creations.

Wait for it to be ACTIVE before proceeding. You can check its status using [AWS console](https://console.aws.amazon.com/gamelift/container-fleets) or [DescribeContainerFleet API](https://docs.aws.amazon.com/gamelift/latest/apireference/API_DescribeContainerFleet.html)

# Usage
Your executable will be started when a game session is created.

## Create Game Session

You can create a game session directly by calling [CreateGameSession API](https://docs.aws.amazon.com/gamelift/latest/apireference/API_CreateGameSession.html).
Example use of AWS CLI to create a game session:

### Managed EC2 and Managed Container Fleets
```bash
aws gamelift create-game-session \
--fleet-id <FLEET_ID> \
--game-properties '[{"Key": "exampleProperty", "Value": "exampleValue"}]' \
--maximum-player-session-count 3 \
--location us-west-2 \
--region us-west-2
```

### Anywhere Fleets
For Anywhere fleets, use the custom location name started with "custom-" for `location` field.

```bash
aws gamelift create-game-session \
--fleet-id <FLEET_ID> \
--game-properties '[{"Key": "exampleProperty", "Value": "exampleValue"}]' \
--maximum-player-session-count 3 \
--location custom-<LOCATION> \
--region us-west-2
```

You can pass customized game properties to your server executable. See [Game Server Arguments](#game-server-arguments) for details.

After receiving the create-game-session call, Amazon GameLift Servers will inform the wrapper to launch the game server executable.

Note:

- Alternatively, you can use Amazon GameLift Servers game session placement feature with [StartGameSessionPlacement](https://docs.aws.amazon.com/gamelift/latest/apireference/API_StartGameSessionPlacement.html), which uses the FleetIQ algorithm and queues to optimize the placement process.

- You can use [TerminateGameSession API](https://docs.aws.amazon.com/gamelift/latest/apireference/API_TerminateGameSession.html) to terminate a game session.

# Metrics

The Game Server Wrapper supports collecting and publishing telemetry metrics from the managed Amazon GameLift Servers host to
AWS services for monitoring and observability. For detailed setup and usage instructions, see [METRICS.md](./metrics/METRICS.md).

![Telemetry Metrics on Amazon Grafana Dashboard](./metrics/telemetry_metrics.png)


# Appendix
## Game Server Arguments
When starting the game server the wrapper is able to provide arguments from configuration and game properties.

The following information can be mapped as an argument to the game server.
```shell
DNSName                    # The DNS identifier assigned to the instance that is running the game session.
FleetId                    # A unique identifier for the fleet that the game session is running on.
GamePort                   # The port number for the game session. To connect to a GameLift game server, an app needs both the IP address and port number.
GameProperties             # A set of custom properties for a game session.  It is in JSON syntax, formatted as a string.
GameSessionData            # A set of custom game session properties, formatted as a single string value.
GameSessionId              # A unique identifier for the game session.
GameSessionName            # A descriptive label that is associated with a game session. Session names do not need to be unique.
IpAddress                  # The IP address of the game session. To connect to a GameLift game server, an app needs both the IP address and port number.
LogDirectory               # Path for the session logs example : /local/game/logs/run_00a42edd-2d01-432e-a0fe-ecd6302ac8bc
MatchmakerData             # Information about the matchmaking process that was used to create the game session. It is in JSON syntax, formatted as a string.
MaximumPlayerSessionCount  # The maximum number of players that can be connected simultaneously to the game session.
```

In addition, game properties from the create-game-session API calls can be mapped as arguments.

Example of configuration of arguments:

```yaml
      defaultArgs:
        - arg: "--port"
          val: "{{.GamePort}}"
          pos: 0
        - arg: "--ipAddress"
          val: "{{.IpAddress}}"
          pos: 1
        - arg: "--gameSessionId"
          val: "{{.GameSessionId}}"
          pos: 2
```

## Metrics
The Game Server Wrapper support collecting and publishing telemetry metrics from the managed Amazon GameLift Servers host to
AWS services for monitoring and observability. For detailed setup and usage instructions, see [METRICS.md](./metrics/METRICS.md).



## Server SDK integration comparison against game server wrapper
The game server wrapper automatically calls some methods from the server SDK for Amazon GameLift servers. To take full advantage of all of the methods, game servers must integrate with the server SDK instead of the game server wrapper.

The following table shows which server SDK methods are utilized by the game server wrapper.

| Method                            | Server SDK Integration | Game Server Wrapper | Wrapper Notes                                             |
|-----------------------------------|------------------------|---------------------|-----------------------------------------------------------|
| GetSdkVersion                     | ✅                     | ❌                   |                                                           |
| InitSDK                           | ✅                     | ✅                   | Called during wrapper initialisation                      |
| ProcessReady                      | ✅                     | ✅                   | Called during wrapper initialisation                      |
| ProcessEnding                     | ✅                     | ✅                   | Called when game terminates                               |
| ActivateGameSession               | ✅                     | ✅                   | Called during OnStartGameSession                          |
| UpdatePlayerSessionCreationPolicy | ✅                     | ❌                   |                                                           |
| GetGameSessionId                  | ✅                     | ❌                   |                                                           |
| GetTerminationTime                | ✅                     | ❌                   |                                                           |
| AcceptPlayerSession               | ✅                     | ❌                   |                                                           |
| RemovePlayerSession               | ✅                     | ❌                   |                                                           |
| DescribePlayerSessions            | ✅                     | ❌                   |                                                           |
| StartMatchBackfill                | ✅                     | ❌                   |                                                           |
| StopMatchBackfill                 | ✅                     | ❌                   |                                                           |
| GetComputeCertificate             | ✅                     | ❌                   |                                                           |
| GetFleetRoleCredentials           | ✅                     | ❌                   |                                                           |
| Destroy                           | ✅                     | ✅                   | Used when closing the game server, after and error occurs |

Note:
- Although the StartMatchBackfill server SDK API is not supported, you can still use [StartMatchBackfill](https://docs.aws.amazon.com/gamelift/latest/apireference/API_StartMatchBackfill.html) AWS SDK API for matchmaking backfill. However, due to lack of support for player session management on the game server side, using Amazon GameLift Servers FlexMatch backfill is not recommended.
- Instead of using GetFleetRoleCredentials, you can use shared credentials file to get fleet role credentials from your server, see details: https://docs.aws.amazon.com/gamelift/latest/developerguide/gamelift-sdk-server-resources.html.
