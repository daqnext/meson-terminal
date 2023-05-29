cbowREADME.md

# node-peer

# How to use

#### 1."default_" application:

##### default_ is the main program

#### 1.1 run default program with {config_name}

##### ```go run . --conf={config_name}``` // will use the {config_name}.toml inside configs folder

##### ```go run .```  // just use defalut.toml

#### 2."config" application:

##### config is the program used to show or set config file

#### 2.1 set config

##### ```go run . --conf={config_name} config set ...```

##### ```go run . config set ...```   //using default.toml

#### 3. log

#### 3.1 show all logs

##### ```go run . log```

#### 3.2 only show error logs : [error,panic,fatal]

##### ```go run . log --only_err=true```

#### 4. "api" application:

##### 4.1 generate the api documents

##### ```go run . gen_api```

## Running process

```
1.entry -> main.go
2.basic logger is initialized 
3.cmd/cmd.go ->ConfigCmd() is called
4.read the related config file
5.--> go to cmd application "config"|"default_"|"log"
```

## build
```
./auto-build.sh
```
