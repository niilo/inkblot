'use strict';

angular.module('inkblot.about')

.config(function ($stateProvider) {
    $stateProvider.state('about', {
        url: '/about/',
        views: {
            "main": {
                controller: 'AboutCtrl',
                templateUrl: 'about/about.tpl.html'
            }
        },
        data: {
            pageTitle: 'About'
        }
    });
})

;
