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
  'recipe[mongodb::default]',
  'recipe[java]',
  'recipe[n-and-nodejs]',
  'recipe[protractor-selenium-server]',
  'recipe[protractor-selenium-server::services]',
]
