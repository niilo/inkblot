#
# Environment definition for a local Vagrant VM.
#

default_attributes(
  'locales' => {
    'available' => 'en_US.utf8',
    'default' => 'en_US.utf8'
  },
  'java' => {
    'install_flavor' => 'oracle',
    'jdk_version' => '8',
    'oracle' => {
      'accept_oracle_download_terms' => true
    }
  },
  'go' => {
    'version' => '1.2.2'
  },
  'n-and-nodejs' => {
    'n' => {
      'version' => '1.2.1'
    },
    'nodejs' => {
      'version' => 'stable'
    }
  },
  'protractor-selenium-server' => {
    # It takes a long time to install Firefox and Chromium via packages on a
    # bare server. It requires many supporting packages.
    'browser-install-timeout' => 1200,
    'selenium' => {
      'install-dir' => '/usr/local/share/selenium',
      'log-dir' => '/var/log/selenium',
      'version' => '2.42.2',
    },
    'xvfb' => {
      'display' => '10',
      'resolution' => '1024x768x24'
    }
  }
)