# Hyprspace Relay
A Containerized Hyprspace Relay

## What is a Hyprspace Relay?
I initially started working on this idea a few years ago in my freshman year of college. At the time, I wanted to run a Nextcloud instance from a server under my bed in my dorm room, but surprise...the university wasn't going to let me forward a port on their network or give me a public ip address. For a while I used a VPN running on a small cloud instance, but that required all of my devices to be always connected to the VPN and I couldn't easily share files with anyone of my friends. So somehow I needed to use my cloud instance to relay information from the web to the Nextcloud server on the VPN.... On my second iteration of the system I started doing that with just a couple of handy iptables rules and a Wireguard server. Even after that system though, I've kept dreaming of something even better, something where I could use a small cloud vm to be the gateway for any number of instances hosting services on local machines without public ip addresses.

Enter, the Hyprspace Relay! At its core, the relay is a containerized Hyprspace instance set up to forward data from any port you specify to a Hyprspace client machine connected on the other side. The relay can either be used in a single portal setup (where your cloud vm is just porting anything sent to it to one client), or a multi-portal setup (where we can use a nginx-reverse-proxy container to forward requests intended for different domains/subdomains to different portals and thus different machines.)

## How is the Hyprspace Relay Different From Running Wireguard in a Container?
Even running Wireguard in a container you'll still need to give it permission to modify the network interfaces on the host. This typically means that you'll have to grant the container access to `NET_ADMIN` and the `SYS_MODULE` and have the Linux Kernel headers installed on the host.

By contrast Hyprspace virtualizes its internal network interfaces so that it can run **completely permissionless**. The operating system doesn't even know a Hyprspace instance is running in the application! All it sees are encrypted libp2p TCP packets entering and exiting like any other application on your computer. From the perspective of running Hyprspace in a Docker Container this is great because it doesn't require any special permissions or any additional packages to be installed on the host.

## Use Cases:
##### A Digital Nomad
I often times to use this system when travelling, if I'm staying in a rental or hotel and want to try something out on a Raspberry Pi I can plug the Pi into the location's router or ethernet port and then just ssh into the system using the same-old `travel-lab.domain.tld` without having to worry about their NAT or local firewall.

##### A Privacy Advocate
Honestly, I even use this system when I'm at home and could port forward from my router to my local infrastructure. Using a Hyprspace Relay however, I don't have to tell every dns server my public home ip address. Instead everyone seeing my infrastructure sees the public ip address of a disposable cloud vm. 

If anyone else has some use cases please add them! Pull requests welcome!

## Getting Started:
### On Your Cloud Instance
1. Create a Virtual Machine on AWS, DigitalOcean, Google Cloud, OVH, etc...
2. Download and Install Docker. Depending on your chosen Linux distribution the instructions will be different so I'll link to the [docker docs here](https://docs.docker.com/engine/install/).
3. (Recommended) Install Docker-Compose from [here](https://docs.docker.com/compose/install/).
4. `sudo systemctl enable --now docker`
5. `sudo docker pull ghcr.io/hyprspace/relay`
6. On your local machine run `sudo hyprspace init relay --config ./relay.yaml`
7. This will generate a config that we can use for our relay.
8. Copy one of the docker compose examples from below into a text editor and replace the fields marked with [[]] with the fields within the configuration file. (Remove the brackets so that it's just the values by the way.)
9. Change the `RELAY_CLIENT_PORTS` and the exposed ports to the ports that you want to relay.
10. If you want to use a different port for the wireguard server, change that and make note of it.
11. Copy the example you just changed into a new file called `docker-compose.yml` on your cloud instance.
12. On your local machine install make sure you've [installed the Hyprspace application](https://github.com/hyprspace/hyprspace).
13. Now run `sudo hyprspace init hs0` to initialize your local config.
14. After the config has been initialized open up the resulting file by typing `sudo nano /etc/hyprspace/hs0.yaml`.
15. Transfer your local node's ID to the peers env variable section on your cloud instance. And similarly set your cloud VM's ID in the "peers" section of your local config.
16. Now we'll need to set the relay's discovery key to be the same as the discovery key of our local machine.
17. Now we're ready to bring up the relay on our cloud vm by typing `docker compose up -d`
18. Now on our local machine we can run `sudo hyprspace up hs0` to bring up the local interface.
19. Give the system a few seconds to find each other and that's it! The requests sent to your relay cloud instance will now automatically be forwarded over Hyprspace to your local server.

## Examples:
### Single Relay Docker-Compose
``` yaml
version: '3'
services:
  portal:
    image: ghcr.io/hyprspace/relay
    restart: always
    environment:
      - RELAY_ID=[[INSERT YOUR PORTAL'S ID HERE]]
      - RELAY_ADDRESS=[[INSERT YOUR RELAY'S INTERNAL IP ADDRESS HERE]]
      - RELAY_PORT=8001
      - RELAY_PRIVATEKEY=[[INSERT YOUR PORTAL'S PRIVATE KEY HERE]]
      - RELAY_CLIENT_IDS=[[INSERT YOUR CLIENT'S PUBLIC KEY HERE]]
      - RELAY_CLIENT_PORTS=80,443
      - RELAY_CLIENT_IPS=10.0.0.2
    # Insert Ports to Relay Plus Server Port.
    ports:
      - 8001:8001
      - 443:443
      - 80:80
```

## License

Copyright 2021-2022 Alec Scott <hi@alecbcs.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
