#
# A role to set up a server ready to run Protractor with headless browsers.
#

run_list [
  # Third party cookbooks. Ensure that apt runs first to force an update.
  'recipe[apt]',
  'recipe[vim]',
  'recipe[locales]',
  'recipe[ntp]',
  'recipe[curl]',
  'recipe[golang]',
  'recipe[docker]',
  'recipe[n-and-nodejs]',
  'recipe[mongodb::default]',
  'recipe[java]',
  'recipe[protractor-selenium-server]',
  'recipe[protractor-selenium-server::services]',
]
