Inkblot
=======

Inkblot will be blogish application for content and social centric publishing and collaboration with readers. Currently it's only technology demo for GO (golang) and AngularJS. Great deal of if's and for loops are missing that this would be called even a beta.


Development
-----------

Inkblot comes with Vagrant development environment. It includes:
* golang
* nodejs + npm + gulp + bower
* mongodb
* git + hg + bzr
* plus some extras (check from Vagrantfile)

### Installation

Download & install Virtualbox + Vagrant:
* https://www.virtualbox.org/
* http://www.vagrantup.com/downloads.html

Install Librarian-Chef:

```sh
$ sudo gem install librarian-chef
```

Install Omnibus plugin for Vagrant:

```sh
$ vagrant plugin install vagrant-omnibus
```

Clone project:

	git clone git@github.com:niilo/inkblot.git


Start and provision development VM (will take several minutes):

```sh
$ cd inkblot
$ librarian-chef install
$ vagrant up
```

If everything seems to be fine connect to newly created instance

	vagrant ssh

For performance reasons ./ directory is synced to virtual instance "/app" with rsync. So to get live updates you need to run rsync on host, so open new terminal window on host machine at same folder where project is and start syncing with

```sh
$ vagrant rsync-auto
```

Then return to guest terminal (where vagrant ssh is running) and run some checks to validate that everything is installed

```sh
$ go version
$ node --version
$ npm --version
$ mongo --version
```

In case of failure : SCREAM! and then google.com

### Running

Your local folder is mounted to "/app" folder. So you can edit files with your favorite editor in your normal operating system and compile & run go + gulp on Vagrant instance.

Let's start inkblot-back application first:

```sh
$ cd /app/api/server
$ go get
$ go build && ./server --conf=inkblot.cfg
```

"go get" get's fetches all needed external dependencies and this is needed only first time and if new dependencies are used in project, so it's like npm but it fetches sources and compiles to binary. "go build && ./server --conf=inkblot.cfg" builds inblot server binary from source and starts server (this will be blazing fast). 

Then we need another ssh shell to same image so start new console and run "vagrant ssh" on project folder. Then we're ready to start Angular application first time:

```sh
$ cd /app/frontend
$ npm update && bower update
$ gulp watch
```

Eventually it will start and you can browse to http://inkblot.vcap.me:4000/

Then you might want to write some nice comments to that nice story. Later on you might want to reply, hate or like some comments.