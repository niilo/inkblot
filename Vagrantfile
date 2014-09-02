# -*- mode: ruby -*-
# vi: set ft=ruby :

VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|

  config.vm.hostname = "inkblot"

  config.vm.define :inkblot do |t|
  end
  
  config.vm.synced_folder "./", "/app/"
  
  #config.vm.network :private_network, ip: "10.0.1.77", :netmask => "255.255.0.0"
  config.vm.network :forwarded_port, host: 4000, guest: 4000
  config.vm.network :forwarded_port, host: 1234, guest: 1234
  config.vm.network :forwarded_port, host: 9999, guest: 9999

  config.vm.provider :virtualbox do |vb|
    vb.name = "inkblot"
    vb.customize ["modifyvm", :id, "--cpus", "2", "--memory", 2048, "--natdnshostresolver1", "on"]
  end

  config.vm.box = "ubuntu64-14.04-trusty"
  config.vm.box_url = "http://cloud-images.ubuntu.com/vagrant/trusty/current/trusty-server-cloudimg-amd64-vagrant-disk1.box"

  config.omnibus.chef_version = :latest
  
  config.vm.provision :shell, :inline => "sudo locale-gen en_US.UTF-8"

  # Enable provisioning with chef solo, specifying a cookbooks path, roles
  # path, and data_bags path (all relative to this Vagrantfile), and adding
  # some recipes and/or roles.
  config.vm.provision :chef_solo do |chef|
    chef.cookbooks_path = "./cookbooks"
    chef.roles_path = "./chef/roles"
    chef.environments_path = "./chef/environments"
    chef.environment = "local-vagrant"
    chef.add_role "inkblot"
  end

  config.vm.provision :shell, :inline => "sudo chown -R vagrant /opt/go"
  config.vm.provision :shell, :inline => "npm install -g gulp grunt-cli bower phantomjs http-server express"
  
end