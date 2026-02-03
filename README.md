# JPM - JVM Package Manager

Website : [ jpmhub.org ](https://www.jpmhub.org/)

<a href="https://www.buymeacoffee.com/jpmhub" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/v2/default-yellow.png" alt="Buy Me A Coffee" style="height: 60px !important;width: 217px !important;" ></a>

![JPM Logo](https://aws-ca-central-1-501301757139-newlambda-pipe.s3.ca-central-1.amazonaws.com/logo2.png)


A simple and efficient build tool and package manager for Java and Kotlin projects. JPM provides a better approach to managing dependencies, building, running, and testing Java/Kotlin applications, inspired by npm.

## Table of Contents

- [Setup JPM Environment](#setup-jpm-environment)
- [Quick Start](#quick-start)
- [Project Structure](#project-structure)
- [Configuration](#configuration)
- [Commands](#commands)
- [Dependencies](#dependencies)
- [Scripts](#scripts)
- [String Substitution](#string-substitution)


## Quick Start
1. **Install JPM**
    <br>SDKMAN! (comming soon)
    ```bash
    # not yet on sdkman but coming soon
    sdk install jpm
    ```
    linux / mac 
    ```bash
    curl -L -o s.sh https://sh.jpmhub.org && sh s.sh
    ```
    windows
    ```bash
    cmd /c "curl -L -o s.cmd https://cmd.jpmhub.org && s.cmd"
    ```
2. **Initialize a new project:**
   ```bash
   # this will create app/App.java
   jpm init 
   # or , this will create a src/main/java/com/example/App.java and initialize a git repo
   jpm init src/main/java/com.example.App -git
   ```

3. **Compile and run:**
   ```bash
   jpm start
   ```

## Project Structure

A typical JPM project structure:

```bash
my-project/
├── package.yml          # Project configuration
├── app/                 # Source code
│   └── App.java
├── tests/               # Test files
│   └── TestApp.java
├── jpm_dependencies/    # Dependencies
│   └── tests/           # Test dependencies
│   └── execs/           # Exec dependencies
├── out/                 # Compiled classes
└── .gitignore           # Git ignore file
```

## Configuration

### package.yml

The main configuration file is an example of a package.yml that defines your project:

```yaml
main: com.example.MyApp
package: example
version: 1.0.0
language: java
src: .
description: an example project
env: .env
envs:
  dev: .env
  test: .test.env
  prod: .prod.env
  custom_env: .custom_env.env
scripts:
  start: jpm compile && jpm run
  dev: jpm watch "(src/**)" "jpm start"
  clean: rm -rf out/*
dependencies:
  - org.apache.commons commons-lang3:latest
  - org.openjfx javafx-control
repos:
  - deafult: https://repo1.maven.org/maven2/
excludes:
  - commons-lang3
classifiers:
  javafx-control: linux
args:
  java: -Xmx512m
  javac: -source 17 -target 17
  kotlinc: -no-stdlib
  junit: --reports-dir=../reports
  hotswap: autoHotswap=false
```

### Configuration Fields

- **package**: Project package name (required)
- **description**: Project description
- **main**: The main class to run (e.g., `com.example.MyApp`)
- **version**: Project version
- **env**: .env file location (you can get the values of keys with `System.getenv("KEY")`)
- **envs**: same thing as env, but has priority if defined
  - `ENV=custom_env jpm run` here we can run jpm with a custom env file
  - `prod` is the env file used when bundling
  - `dev` is the env file used when running `jpm compile` and `jpm run` in any case
  - `test` is the env file used when running `jpm test` in any case
- **src**: Project source dir (usually just . or src/main/java)
- **language**: Programming language (`java` or `kotlin` or `java,kotlin`)
- **scripts**: Custom commands (see [Scripts](#scripts))
- **dependencies**: Project dependencies (see [Dependencies](#dependencies))
- **excludes**: list of excludes package names or artifactIDs
- **classifiers**: map of classifiers keys can be package name, artifactIDs, groupIDs or "*"
  - **org.openjfx: win** -> example of classified groupID
- **repos**: Repository configurations
  - **mvn: https://repo1.maven.org/maven2/** -> example of repository, preppend all dependencies with `mvn` to install from that specific repository, `default` means no need to preppend anything.
- **args**: Command-line arguments for different operations
  - `java` -> executes on the java command
  - `javac` -> executes on the javac command
  - `kotlinc` -> executes on the kotlinc command
  - `junit` -> args are added to junit5 jar
  - `<any-exec-dependencies>` -> args added to scripts of exec dependencies

## Commands

### Upgrading

#### `jpm upgrade`
execute the upgrade script to get latest version

```bash
jpm upgrade
```

### Project Management

#### `jpm init [project_name] [options]`
Initialize a new Java project.

```bash
# Basic initialization
jpm init 

# Basic initialization of a kotlin app
jpm init new-app.App -kt

# With Git repository 
jpm init new-app -git

# With Dockerfile
jpm init src/main/kotlin/com.example.myapp.MyApp -kt -docker
```

#### `jpm create <template>`
Create a project from a template.

```bash
# looks up jpm's repository for simple-spring-app template
jpm create simple-spring-app

# looks up the working dir for simple-spring-app.yml template (you can create your own templates)
jpm create -yml simple-spring-app.yml
```

### Development

#### `jpm compile`
Compile the source code under the package source dir and package dir.

```bash
jpm compile
```

#### `jpm run [options] [...appArgs]`
Run the compiled application.

```bash
# Basic run
jpm run 

# Basic run with args
jpm run appArg1 appArg2

# With watch mode (hot reloading)
jpm run -hot

# Omitting the -hot args and applying your application args
jpm run -hot _ _ appArg1 appArg2

# With custom arguments on each file .java or .html change
jpm run -hot "(src/**.java|src/**.html)" "echo reloading" appArg1

```

#### `jpm watch [pattern] [command]`
Watch for file changes and execute commands.

```bash
# Watch all files under src and compile
jpm watch 

# Watch all Java files and run tests
jpm watch "(src/**.java)" "jpm test"

# Watch specific pattern and start
jpm watch "(src/com/example/**.java)" "jpm start"
```

#### `jpm args [key]`
yields the arguments written in package.yml for the key specified<br>
very helpful fro executable dependencies

```yaml
# in package.yml
args:
  javac: -parameters
  dex: -path .
```

```bash
# in the shell
$ jpm args "javac"
-parameters

$ jpm args "dex"
-path .
```

### Testing

#### `jpm test`
Run tests using JUnit.

```bash
jpm test
```

### Dependency Management

#### `jpm install`
Install all dependencies from `package.yml`.

```bash
jpm install

# force a re-resolving of all dependencies (delete jpm_dependencies to re-install)
jpm install -f
```

#### `jpm install <dependency> [scope]`
Install a specific dependency.

```bash
# Install from JPM repository
jpm install my-library

# Install from Maven repository if maven is set as default repository
jpm install org.apache.derby derby:10.17.1.0

# Install from custom repository if alias was set in package.yml
jpm install my-repo org.example library:1.0.0

# Install raw JAR file
jpm install raw https://example.com/library.jar

# Install with extraction
jpm install raw -x https://example.com/library.zip

# Install with scope
jpm install my-library:1.0.0 test
jpm install concurrently:1.0.0 exec
```

#### `jpm install -repo <alias>:<url>`
Adds a new repository.

```bash
jpm install -repo mvn:https://repo1.maven.org/maven2/
```

### Building

#### `jpm bundle [options]`
Create a JAR bundle.

```bash
# Create basic JAR
jpm bundle 

# Create fat JAR (all dependencies included, not yet surported)
jpm bundle -fat

# Create a executable JAR with scripts
jpm bundle -exec

# Create a executable JAR with scripts, creates dependencies.json and a MD file.
jpm bundle -publish

# Create a executable JAR with scripts, creates dependencies.json and a MD file.
# keeps the classifiers you've set in you project in dependencies.json
jpm bundle -publish -keep-classifiers
```

#### `jpm setup <component>`
Setup JPM components.

```bash
# Setup Java for HotSwap Agent
jpm setup -java

# Setup Kotlin
jpm setup -kotlin

# Setup JUnit
jpm setup -junit

# Setup HotSwap Agent
jpm setup -hotswap

# toggle verbosity
jpm setup -v

# install git (especially for windows)
jpm setup -git

# installs openjdk
jpm setup -jar

# setup jpx to execute you exec dependencies on the cli
jpm setup -jar
```

### Custom Scripts

You can run custom scripts defined in `package.yml`:

```bash
jpm <script-name>
```

if you have a script called `run@` in your `package.yml`, it overrides the default `jpm run` to your custom command.
example for `jpm init`

```bash
$ jpm init
Overriding: 'jpm init' for 'jpm init@'
```

You can append the rest of the command provided to jpm by adding `...args@` to the script
this way you can still use jpm install with args, but it'll also download `image.png`
```yaml
scripts:
  install@: |
    wget -O resources/image.png https://openimage.com/images/image.png
    jpm install ...args@
```

## Dependencies

JPM supports multiple types of dependencies:

### Maven Dependencies

```yaml
dependencies:
  # installs derby for the project
  - org.apache.derby derby:latest
  # installs springboot's test libraries in a test scope
  - org.springframework.boot spring-boot-starter-test test
```

### JPM Dependencies

```yaml
dependencies:
  # installs my-devDependecny in the exec scope, it makes it executable in scripts 
  - my-devDependency:1.0.0 exec
  # installs neutron from jpm's repository
  - neutron
```

### Raw Dependencies

```yaml
dependencies:
  - raw https://example.com/library.jar exec # custom jar now available for execution
  - raw -x https://example.com/library.tar.gz # tar.gz extraction support
  - raw -x https://example.com/library.zip # zip extraction support
```

### Repository Dependencies

```yaml
dependencies:
  # custom POM repository using the alias "my-repo", you have to preppend this alias to use the repository
  - my-repo org.example library:1.0.0
repos:
  - my-repo: https://my-custom-repo.com/repository/
```

### Dependency Scopes

- **test**: Available only for testing
- **exec**: Executable dependencies to run with `jpx` 

## Scripts

Define custom scripts in your `package.yml`:

```yaml
port: "8090"
scripts:
  dev: jpm run -hot _ _ {{ port }}
  start: jpm compile && jpm run {{ port }}
  clean:deep: |
    echo cleaning
    rm -rf out
    rm -rf dist
    rm -rf jpm_dependencies
  build: jpm compile && cp ressourses/* out/
```
You can override jpm scripts by appending @ in front of the jpm script in your custom scripts:

```yaml
scripts:
  compile@: jpm compile && cp ressourses/* out/
```

You can omit the execution of an overriden jpm scripts by appending ! when calling the script:

```yaml
scripts:
  compile@: jpm compile && cp ressourses/* out/
  omit: jpm compile! && cp ressourses/* out/     # this will execute the normal compile, not the overriden one
```

After overriding a script you can still append the args that was called with the script by adding `...args@`

```yaml
scripts:
  # this will replace ...args@ with the actuall args that was called while overriding the script
  install@: CLASSIFIER=javadoc jpm install ...args@ 
```

## String substitution

1. **You can substitute an arbitrary string in the package.yml**
    ```yaml
    springVerison: 3.5.5
    dependencies:
      - org.springframework.boot spring-boot-starter-web:{{ springVerison }}
    ```

2. **You can substitute an arbitrary key=value in the `.env` file**
    ```yaml
    env: .env
    scripts:
      start: jpm compile && jpm run --server.port={{ env.PORT }}
    ```

3. **You can substitute an arbitrary environment variable with `ENV.` :**
    <br>shell:
    ```bash
    $ JFX=mac jpm install
    ```
    package.yml:
    ```yaml
    classifiers:
      org.openjfx: "{{ ENV.JFX }}"
    ```
4. **You can substitute platform infromation with `jpm.` :**
    <br>package.yml:
    ```yaml
    scripts:
      print-info: echo {{ jpm.OS }} {{ jpm.ARCH }} {{ jpm.OS-ARCH }}
    ```
    shell:
    ```bash
    $ jpm print-info

      linux amd64 linux-amd64
    ```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License. 