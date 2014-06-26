'use strict';

angular.module('inkblot.commentsDirective', [])
    
.directive('inkblotComments', ['commentsFactory', function (commentsFactory) {
    return {
        restrict: 'AE',
        scope: {},
        link: function (scope, elem, attrs) {
            scope.openEditor = function (index) {
                scope.editMode = true;
                if (index !== undefined) {
                    scope.commentText = commentsFactory.get(index).content;
                    scope.index = index;
                } else {
                    scope.commentText = undefined;
                }
            };
            scope.save = function () {
                if (scope.commentText !== "" && scope.commentText !== undefined) {
                    var comment = {};
                    comment.title = scope.commentText.length > 10 ? scope.commentText.substring(0, 10) + '. . .' : scope.commentText;
                    comment.content = scope.commentText;
                    comment.id = scope.index !== -1 ? scope.index : localStorage.length;
                    scope.comments = commentsFactory.put(comment);
                }
                scope.restore();
            };


            scope.restore = function () {
                scope.editMode = false;
                scope.index = -1;
                scope.commentText = "";
            };

            var editor = elem.find('#editor');

            scope.restore();

            scope.comments = commentsFactory.getAll();

            editor.bind('keyup keydown', function () {
                scope.commentText = editor.text().trim();
            });

        },
        templateUrl: 'directives/commentsDirective.tpl.html'
    };
}])

.factory('commentsFactory', [function () {
    return {
        put: function (comment) {
            localStorage.setItem('comment' + comment.id, JSON.stringify(comment));
            return this.getAll();
        },
        get: function (index) {
            return JSON.parse(localStorage.getItem('comment' + index));
        },
        getAll: function () {
            var comments = [];
            for (var i = 0; i < localStorage.length; i++) {
                if (localStorage.key(i).indexOf('comment') !== -1) {
                    var comment = localStorage.getItem(localStorage.key(i));
                    comments.push(JSON.parse(comment));
                }
            }
            return comments;
        }
    };
}]);