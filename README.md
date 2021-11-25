# Meson Network Terminal

Meson Network is a bandwidth trading platform built on blockchain.

## Demo video

[![3 Minutes To Start Mining Meson Network](https://img.youtube.com/vi/jHrVCpuREqk/maxresdefault.jpg)](https://youtu.be/jHrVCpuREqk)

![](https://www.youtube.com/s/desktop/c3755e48/img/favicon.ico)
[3 Minutes To Start Mining Meson  Network~
](https://www.youtube.com/watch?v=jHrVCpuREqk) 

## Deploy Meson Network

### Linux

```
wget 'https://coldcdn.com/api/cdn/f2cobx/terminal/v2.5.1/meson-linux-amd64.tar.gz'    # download the terminal package
tar -zxf meson-linux-amd64.tar.gz         # unzip the package
cd ./meson-linux-amd64                    # install the app as service
sudo ./meson service-install              # input your token, port and space provide
sudo ./meson service-start                # start the app
sudo ./meson service-status               # wait about 1 minutes and check status
sudo ./meson service-stop                 # to stop meson network service
sudo ./meson service-remove               # to remove meson network application
```

### Windows

```
wget 'https://coldcdn.com/api/cdn/f2cobx/terminal/v2.5.1/meson-windows-amd64.zip'    # download the terminal package
unzip meson-windows-amd64.zip             # unzip the package
cd ./meson-windows-amd64 && ./meson.exe   # run the app
```

### Mac

```
wget 'https://coldcdn.com/api/cdn/f2cobx/terminal/v2.5.1/meson-darwin-amd64.tar.gz'    # download the terminal package
tar -zxf meson-darwin-amd64.tar.gz        # unzip the package
cd ./meson-darwin-amd64                   # install the app as service
./meson service-install                   # input your token, port and space provide
./meson service-start                     # start the app
./meson service-status                    # wait about 1 minutes and check status
./meson service-stop                      # to stop meson network service
./meson service-remove                    # to remove meson network application
```

## Reference

- Meson Network Documentation: https://docs.meson.network/