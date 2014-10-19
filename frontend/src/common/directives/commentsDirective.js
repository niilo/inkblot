'use strict';

angular.module('inkblot.commentsDirective', ['restangular'])

.config(function (RestangularProvider) {
  RestangularProvider.setBaseUrl('http://inkblot.vcap.me:1234');
})
    
.directive('inkblotComments', ['commentsFactory', function (commentsFactory) {
    return {
        restrict: 'AE',
        scope: {},
        link: function (scope, elem, attrs) {
            scope.openEditor = function (index) {
                scope.editor = true;
                scope.editMode = true;
                if (index !== undefined) {
                    scope.commentText = commentsFactory.get(index).text;
                    scope.index = index;
                } else {
                    scope.commentText = undefined;
                }
            };
            
            scope.reply = function (cid, comment) {
                scope.editor = true;
                scope.editMode = true;
                var text = '\n[quote][to]' + cid + '[/to]' + comment + '[/quote]\n';
                console.log(text);
                if (scope.commentText === undefined) {
                    scope.commentText = '';
                }
                scope.commentText += text;
                angular.element('#editor').trigger('focus');
            };
            
            scope.openAbuseMessageEditor = function (index, text) {
                scope.editor = true;
                scope.reportAbuse = true;
                if (index !== undefined) {
                    //scope.commentText = commentsFactory.get(index).text;
                    scope.index = index;
                    scope.abuseAboutContent = text;
                } else {
                    scope.commentText = undefined;
                }
            };
            
            scope.save = function () {
                if (scope.commentText !== "" && scope.commentText !== undefined) {
                    var comment = {};
                    comment.text = scope.commentText;
                    comment.id = scope.index !== -1 ? scope.index : localStorage.length;
                    commentsFactory.put(scope, comment);
                }
                scope.restore();
            };

            scope.likeComment = function (index) {
                commentsFactory.like(index);
            };
            
            scope.hateComment = function (index) {
                commentsFactory.hate(index);
            };
            
            scope.abuse = function () {
                if (scope.commentText !== "" && scope.commentText !== undefined) {
                    var comment = {};
                    comment.text = scope.commentText;
                    comment.id = scope.index !== -1 ? scope.index : localStorage.length;
                    commentsFactory.put(scope, comment);
                }
                scope.restore();
            };
            
            scope.restore = function () {
                scope.editMode = false;
                scope.editor = false;
                scope.reportAbuse = false;
                scope.index = -1;
                scope.commentText = "";
            };

            var editor = elem.find('#editor');

            scope.restore();

            commentsFactory.getAll(scope);

            editor.bind('keyup keydown', function () {
                scope.commentText = editor.text().trim();
            });
            
        },
        templateUrl: 'directives/commentsDirective.tpl.html'
    };
}])

.factory('commentsFactory', ['Restangular', function (Restangular) {
    
    var restComments = Restangular.all('story');
    
    return {        
        put: function (scope, comment) {
            comment.author = 'anonymous';
            comment.published = new Date();
            //localStorage.setItem('comment' + comment.id, JSON.stringify(comment));
            //Restangular.all('story').post(comment);
            restComments.all('kudv8').all('comment').post(comment).then(function(comments) {
                console.log("Object saved OK");
                scope.story = comments;
            }, function() {
                console.log("There was an error saving");
            });
        },
        get: function (index) {
            //return JSON.parse(localStorage.getItem('comment' + index));
            restComments.one('kudv8').one('comments').get();
        },
        getAll: function (scope) {
            /*var comments = [];
            for (var i = 0; i < localStorage.length; i++) {
                if (localStorage.key(i).indexOf('comment') !== -1) {
                    var comment = localStorage.getItem(localStorage.key(i));
                    comments.push(JSON.parse(comment));
                }
            }
            return comments;*/
            Restangular.one('story', 'kudv8').one('comments').get().then(function (comments) {
                //console.log('got comment::' + JSON.stringify(comments));
                scope.story = comments;
            });
        },
        like: function (id) {
            Restangular.one('comment', id).one('like').put().then(function () {
                console.log('liked comment::' + id);
            });
        },
        hate: function (id) {
            Restangular.one('comment', id).one('hate').put().then(function () {
                console.log('hated comment::' + id);
            });
        }
    };
}])

.filter('quote', function () {
    return function (input) {

      String.prototype.convertQuoteTagToHtmlTag = function (quoteTag, htmlTag) {
        var bs = '[' + quoteTag + ']';
        var hs = '<' + htmlTag + '>';
        var be = '[/' + quoteTag + ']';
        var he = '</' + htmlTag + '>';
        var output = this;
        var i = -1;
        while (output.indexOf(be) !== -1) {
          i = output.indexOf(be);
          output = output.substring(0, i) + he + output.substring(i + be.length);
        }
        while (output.indexOf(bs) !== -1) {
          i = output.indexOf(bs);
          output = output.substring(0, i) + hs + output.substring(i + bs.length);
        }
        return output;
      };

      String.prototype.replaceAllwoRegExp = function (from, to) {
        var output = this;
        while (output.indexOf(from) !== -1) {
          var i = output.indexOf(from);
          output = output.substring(0, i) + to + output.substring(i + to.length);
        }
        return output;
      };

      String.prototype.convertLineFeedToHtmlBreak = function () {
        var br = '<br>';
        var output = this;
        return output.replaceAllwoRegExp('\r\n', br).replaceAllwoRegExp('\n', br);
      };

      return input.convertQuoteTagToHtmlTag('quote', 'blockquote').convertQuoteTagToHtmlTag('to', 'cite');

    };
  });