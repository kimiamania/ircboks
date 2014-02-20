ircboksApp.directive('scrollItem',function(){
    return{
    restrict: "A",
    link: function(scope, element, attributes) {
        if (scope.$last){
           scope.$emit("Finished");
       }
    }
   };
});

ircboksApp.directive('scrollIf', function() {
return{
    restrict: "A",
    link: function(scope, element, attributes) {
        scope.$on("Finished",function(){
            var chat_hist = scope.chat_hist;
            if (chat_hist.lastScrollPos === undefined ||
              chat_hist.needScrollBottom === true) {
                var scrollHeight = element[0].scrollHeight;
                element.scrollTop(scrollHeight);
                chat_hist.needScrollBottom = false;
            } else {
              element.scrollTop(chat_hist.lastScrollPos);
            }
        });
    }
   };
  });