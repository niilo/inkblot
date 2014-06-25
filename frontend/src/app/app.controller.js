'use strict';

angular.module('inkblot')

.controller('AppCtrl', function ($rootScope, $scope) {

    // handling UI Bootstrap Collapse plugin
    $scope.isCollapsed = true;

    $rootScope.$on('$stateChangeSuccess', function (event, toState, toParams, fromState, fromParams) {
        if (angular.isDefined(toState.data.pageTitle)) {
            $scope.pageTitle = toState.data.pageTitle + ' | inkblot';
        }
    });
})

;
