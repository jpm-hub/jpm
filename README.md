# JPM - JVM Package Manager

Website : [ jpmhub.org ](https://www.jpmhub.org/)

![JPM Logo](https://aws-ca-central-1-501301757139-newlambda-pipe.s3.ca-central-1.amazonaws.com/logo2.png)


A simple and efficient build tool and package manager for Java and Kotlin projects. JPM provides a streamlined approach to managing dependencies, building, running, and testing Java/Kotlin applications.

## Table of Contents

- [Setup JPM Environment](#setup-jpm-environment)
- [Quick Start](#quick-start)
- [Project Structure](#project-structure)
- [Configuration](#configuration)
- [Commands](#commands)
- [Dependencies](#dependencies)
- [Scripts](#scripts)
- [Examples](#examples)
- [Troubleshooting](#troubleshooting)


## Setup JPM Environment

After building and adding jpm to your PATH, set up the JPM environment with required tools:

```bash
# Setup Java with DCEVM (recommended for hot reloading)
jpm setup -java

# Setup Kotlin compiler
jpm setup -kotlin

# Setup JUnit for testing
jpm setup -junit

# Setup HotSwap Agent hot class reloading
jpm setup -HotSwapAgent
```

## Quick Start

1. **Initialize a new project:**
   ```bash
   jpm init 
   # or , this will create a src/com/example/App.java and initialize a repo
   jpm init src/com.example.App -git
   ```

2. **Compile and run:**
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
env: .env
scripts:
  start: jpm compile && jpm run
  dev: jpm start && jpm watch "(src/**)" "jpm start"
  clean: rm -rf out/*
dependencies:
  - mvn org.apache.commons commons-lang3:latest
repos:
  - mvn: https://repo1.maven.org/maven2/
args:
  run: -Xmx512m
  javac: -source 17 -target 17
  kotlinc: -no-stdlib
  test: #some junit5 command-line arg
  hotswap: autoHotswap=false
```

### Configuration Fields

- **main**: The main class to run (e.g., `com.example.MyApp`)
- **package**: Project package name (required for publishing a package to jpm)
- **version**: Project version
- **language**: Programming language (`java` or/and `kotlin`)
- **scripts**: Custom commands (see [Scripts](#scripts))
- **dependencies**: Project dependencies (see [Dependencies](#dependencies))
- **repos**: Repository configurations
- **args**: Command-line arguments for different operations
  - `run` -> executes on the java command
  - `javac` -> executes on the javac command
  - `kotlinc` -> executes on the kotlinc command
  - `test` -> args are added to junit5 jar
  - `hotswap` -> args added to hotswap-agent

## Commands

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
jpm init src/main/java/com.example.myapp.MyApp -docker
```

#### `jpm create <template>`
Create a project from a template.

```bash
# looks up jpm's repository for simple-spring-app template
jpm create simple-spring-app

# looks up the working dir for simple-spring-app.yml template
jpm create -yml simple-spring-app.yml
```

### Development

#### `jpm compile`
Compile the Java source code.

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

# Omitting the watch args and applying your application args
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

# force a reinstall of all dependencies
jpm install -f
```

#### `jpm install <dependency> [scope]`
Install a specific dependency.

```bash
# Install from JPM repository
jpm install my-library

# Install from Maven repository
jpm install mvn org.apache.derby derby:10.17.1.0

# Install from custom repository
jpm install my-repo org.example library:1.0.0

# Install raw JAR file
jpm install raw https://example.com/library.jar

# Install with extraction
jpm install raw -x https://example.com/library.zip

# Install with scope
jpm install my-library:1.0.0 test
```

#### `jpm install -repo <alias>:<url>`
Add a new repository.

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
```

### System

#### `jpm doctor`
Check JPM installation and dependencies.

```bash
jpm doctor
```

#### `jpm setup <component>`
Setup JPM components.

```bash
# Setup Java with DCEVM
jpm setup -java

# Setup Kotlin
jpm setup -kotlin

# Setup JUnit
jpm setup -junit

# Setup HotSwap Agent
jpm setup -HotSwapAgent

# Setup verbose
jpm setup -verbose
```

### Custom Scripts

You can run custom scripts defined in `package.yml`:

```bash
jpm <script-name>
```

if you have a script called `run@` in your `package.yml`, it overrides the default `jpm run` to your custom command, it only works if no other args was provided to the command.
example for `jpm init`

```bash
$ jpm init
Overriding: 'jpm init' for 'jpm init@'
```

## Dependencies

JPM supports multiple types of dependencies:

### Maven Dependencies

```yaml
dependencies:
  - mvn org.apache.derby derby:latest
  - mvn org.junit.jupiter junit-jupiter:5.8.2 test
```

### JPM Dependencies

```yaml
dependencies:
  - my-library:1.0.0
  - another-package
```

### Raw Dependencies

```yaml
dependencies:
  - raw https://example.com/library.jar
  - raw -x https://example.com/library.tar.gz # tar.gz support
  - raw -x https://example.com/library.zip # zip support
```

### Repository Dependencies

```yaml
dependencies:
  - my-repo org.example library:1.0.0
```

### Dependency Scopes

- **test**: Available only for testing
- **exec**: Executable dependencies to run with `jpmx` (not yet available)
## Scripts

Define custom scripts in your `package.yml`:

```yaml
scripts:
  dev: jpm run -hot
  start: jpm compile && jpm run
  clean: |
    echo cleaning
    rm -rf out
    rm -rf dist
    rm -rf jpm_dependencies
  build: cp ressourses/* out/ && jpm compile
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