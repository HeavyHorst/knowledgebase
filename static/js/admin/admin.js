window.bus = new Vue();
const app = new Vue({
  el: "#site-content",
  data: {
    token: "",

    userProfile: {},

    categories: {},
    articles: [],
    users: [],

    artLimit: 10,
    artOffset: 0,
    artTotal: 0,
    artPaginate: true,

    category: { id: "" },
    article: { id: "" },
    user: { username: "" },

    view: "categories",

    mdconverter: new showdown.Converter(),
    printTime: function(item) {
      var t = moment(item);
      t.locale("de");
      return t.fromNow();
    },
    mdToHtml: function(text) {
      return this.mdconverter.makeHtml(text);
    }
  },
  created: function() {
    bus.$on("refresh-categories", this.updateCategories);
    bus.$on("refresh-articles", this.updateArticles);
    bus.$on("refresh-users", this.fetchAllUsers);
    bus.$on("set-token", this.setToken);
  },
  mounted: function() {
    var that = this;
    this.refreshToken(function() {
      that.loadProfile();
      that.fetchAllCategories();
    });
  },
  updated: function() {
    window.componentHandler.upgradeAllRegistered();

    $("#mainsearch").unbind("keyup");
    var that = this;
    $("#mainsearch").on("keyup", function(e) {
      if (e.keyCode === 27) {
        $(e.currentTarget).val("");
        if (that.view === "categories") {
          that.fetchAllCategories();
        } else {
          that.fetchAllArticles();
        }
      } else if (e.keyCode === 13) {
        if (that.view === "categories") {
          that.searchCategories(e.currentTarget.value);
        } else {
          that.searchArticles(e.currentTarget.value);
        }
      }
    });
  },
  watch: {
    artLimit: function() {
      this.fetchAllArticles();
    },
    artOffset: function() {
      this.fetchAllArticles();
    }
  },
  methods: {
    scrollTop: function() {
      this.$nextTick(function() {
        document.querySelector(".mdl-layout__content").scrollTop = 0;
      });
    },
    removeToken: function() {
      this.token = "";
      localStorage.removeItem("token");
      location.reload();
    },
    setToken: function(token) {
      this.token = token;
      localStorage.setItem("token", token);
    },
    refreshToken: function(callback) {
      var that = this;
      this.token = localStorage.getItem("token");

      if (this.token) {
        var decoded_token = jwt_decode(this.token);
        if (Date.now() / 1000 < decoded_token.exp) {
          $.ajax({
            url: "/api/authorize/refresh",
            type: "GET",
            headers: { Authorization: "Bearer " + that.token },
            success: function(data, status, xhr) {
              that.setToken(data.token);

              if (callback) {
                callback();
              }
            }
          });
        } else {
          this.openLoginDialog(callback);
        }
      } else {
        this.openLoginDialog(callback);
      }
      setTimeout(function() {
        $("body").css("visibility", "visible");
      }, 200);
    },
    loadProfile: function() {
      if (this.token) {
        var that = this;
        var decoded_token = jwt_decode(this.token);
        $.ajax({
          url: "/api/users/" + decoded_token.sub,
          type: "GET",
          headers: { Authorization: "Bearer " + that.token },
          success: function(data, status, xhr) {
            that.userProfile = data;
          }
        });
      }
    },
    incArtOffset: function() {
      if (this.artOffset + this.articles.length < this.artTotal) {
        this.artOffset += this.artLimit;
      }
    },
    decArtOffset: function() {
      if (this.artOffset >= this.artLimit) {
        this.artOffset -= this.artLimit;
      } else {
        this.artOffset = 0;
      }
    },
    setUserView: function() {
      this.view = "users";
      this.scrollTop();
    },
    setArticleView: function() {
      this.view = "articles";
      this.scrollTop();
    },
    setCategoryView: function() {
      this.view = "categories";
      this.scrollTop();
    },
    fetchCategoriesForCategory: function(event) {
      var that = this;
      var id = $(event.target).parents(".mdl-list__item").attr("id");
      var margin = Number($("#" + id).css("margin-left").replace("px", ""));
      var index = 0;
      for (var i = 0; i < this.categories.length; i++) {
        if (this.categories[i].ID == id) {
          index = i;
        }
      }

      if (!$("#" + id).hasClass("is-expanded")) {
        this.fetchCategories("/api/categories/category/" + id, function(json) {
          if (json) {
            var json = json.map(function(elem) {
              elem.margin = margin + 50;
              return elem;
            });

            that.categories.splice.apply(
              that.categories,
              [index + 1, 0].concat(json)
            );
          }
        });
      } else {
        for (var i = index + 1; i < that.categories.length; i++) {
          if (that.categories[i].margin > margin) {
            that.categories.splice(i, 1);
            i--;
          } else {
            break;
          }
        }
      }
      $("#" + id).toggleClass("is-expanded");
    },
    fetchCategories: function(url, callback) {
      var that = this;
      $.ajax({
        url: url,
        type: "GET",
        headers: { Authorization: "Bearer " + that.token },
        success: function(json) {
          if (callback) {
            callback(json);
          }
        }
      });
    },
    fetchAllCategories: function() {
      var that = this;

      this.fetchCategories("/api/categories/category/", function(json) {
        if (json) {
          that.categories = json;
        } else {
          that.categories = [];
        }
        that.setCategoryView();
      });
    },
    updateCategories: function(id) {
      var that = this;

      if (!that.categories) {
        that.categories = [];
      }

      this.fetchCategories("/api/categories/" + id, function(json) {
        var index = 0;
        var found = false;
        var categoryChanged = false;
        json.margin = 0;

        for (var i = 0; i < that.categories.length; i++) {
          if (that.categories[i].ID == id) {
            index = i;
            found = true;
            if (that.categories[i].category != json.category) {
              categoryChanged = true;
            }
          }
          if (that.categories[i].ID == json.category) {
            json.margin = (that.categories[i].margin || 0) + 50;
          }
        }

        if (found && !categoryChanged) {
          that.categories.splice(index, 1, json);
          return;
        } else if (categoryChanged) {
          that.categories.splice(index, 1);
        }

        if (json.category == "") {
          that.categories.push(json);
        } else {
          for (var i = 0; i < that.categories.length; i++) {
            if (that.categories[i].ID == json.category) {
              that.categories.splice(i + 1, 0, json);
              return;
            }
          }
        }
      });
    },
    updateArticles: function(id) {
      var that = this;

      if (!that.articles) {
        that.articles = [];
      }

      this.fetchArticles("/api/articles/" + id, function(json) {
        for (var i = 0; i < that.articles.length; i++) {
          if (that.articles[i].ID == id) {
            that.articles.splice(i, 1, json);
            return;
          }
        }
        that.articles.push(json);
      });
    },
    fetchArticles: function(url, callback) {
      var that = this;
      $.ajax({
        url: url,
        type: "GET",
        headers: { Authorization: "Bearer " + that.token },
        data: { offset: that.artOffset, limit: that.artLimit },
        success: function(json, status, xhr) {
          that.setArticleView();

          if (callback) {
            callback(json, status, xhr);
          }
        }
      });
    },
    fetchArticlesForCategory: function(event) {
      var that = this;
      that.artPaginate = false;
      var id = $(event.target).parents(".mdl-list__item").attr("id");
      this.fetchArticles("/api/articles/category/" + id, function(json) {
        //that.activeCategory = id;
        that.articles = json;
        that.category = { id: id };
      });
    },
    fetchAllArticles: function() {
      var that = this;
      that.artPaginate = true;
      this.fetchArticles("/api/articles", function(json, status, xhr) {
        //that.activeCategory = "";
        that.articles = json;
        that.category = { id: "" };
        that.artTotal = Number(xhr.getResponseHeader("X-Total-Count"));
      });
    },
    fetchAllUsers: function() {
      var that = this;
      $.ajax({
        url: "/api/users",
        type: "GET",
        headers: { Authorization: "Bearer " + that.token },
        success: function(data, status, xhr) {
          that.users = data;
          that.setUserView();
        }
      });
    },
    searchCategories: function(query) {
      var that = this;
      $.ajax({
        url: "/api/categories/search",
        type: "GET",
        headers: { Authorization: "Bearer " + that.token },
        data: { q: query },
        success: function(json) {
          that.categories = json;
          that.setCategoryView();
        }
      });
    },
    searchArticles: function(query) {
      var that = this;
      $.ajax({
        url: "/api/articles/search",
        type: "GET",
        headers: { Authorization: "Bearer " + that.token },
        data: { q: query },
        success: function(json) {
          that.articles = json;
          that.setArticleView();
        }
      });
    },
    openUserDialog: function(id, method) {
      var that = this;
      this.user.username = id;
      this.$nextTick(function() {
        bus.$emit("showUserDialog", that.token, method);
      });
    },
    openLoginDialog: function(callback) {
      this.$nextTick(function() {
        bus.$emit("showLoginDialog", callback);
      });
    },
    openCategoryDialog: function(id) {
      var that = this;
      this.category.id = id;
      this.$nextTick(function() {
        bus.$emit("showCategoryDialog", that.token);
      });
    },
    openArticleDialog: function(id, category) {
      var that = this;
      this.article.id = id;
      this.$nextTick(function() {
        bus.$emit("showArticleDialog", category, that.token);
      });
    },
    changeUserDialog: function(event) {
      var id = $(event.target).parents("tr").attr("id");
      this.openUserDialog(id, "PUT");
    },
    changeCategoryDialog: function(event) {
      var id = $(event.target).parents(".mdl-list__item").attr("id");
      this.openCategoryDialog(id);
    },
    changeArticleDialog: function(event) {
      var id = $(event.target).parents("tr").attr("id");
      this.openArticleDialog(id);
    },
    newUserDialog: function() {
      this.openUserDialog("", "POST");
    },
    newCategoryDialog: function() {
      this.openCategoryDialog();
    },
    newArticleDialog: function() {
      this.openArticleDialog(null, this.category.id);
    },
    deleteArticle: function(event) {
      var that = this;
      var id = $(event.target).parents("tr").attr("id");
      deleteResource("/api/articles/" + id, this.token, function() {
        for (var i = 0; i < that.articles.length; i++) {
          if (that.articles[i].ID == id) {
            that.articles.splice(i, 1);
            return;
          }
        }
      });
    },
    deleteCategory: function(event) {
      var that = this;
      var id = $(event.target).parents(".mdl-list__item").attr("id");
      deleteResource("/api/categories/" + id, this.token, function() {
        for (var i = 0; i < that.categories.length; i++) {
          if (that.categories[i].ID == id) {
            that.categories.splice(i, 1);
            return;
          }
        }
      });
    },
    deleteUser: function(event) {
      var that = this;
      var id = $(event.target).parents("tr").attr("id");
      deleteResource("/api/users/" + id, this.token, that.fetchAllUsers);
    }
  }
});

function deleteResource(resource, token, callback) {
  $.ajax({
    url: resource,
    headers: { Authorization: "Bearer " + token },
    type: "DELETE",
    success: callback
  });
}
