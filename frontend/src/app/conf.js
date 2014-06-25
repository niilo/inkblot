'use strict';

angular.module('inkblot.conf', [])

.constant('conf', {
    api: {
        login  : '/api/login',
        logout : '/api/logout',
        signup : '/api/signup',
        expiry : '/api/expiry'
    }
});
