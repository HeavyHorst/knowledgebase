window.bus = new Vue();
const app = new Vue({
    el: '#site-content',
    data: {
        token: "",

        userProfile: {},

        categories: {},
        articles: [],
        users: [],

        artLimit: 10,
        artOffset: 0,
        artPaginate: true,

        category: { id: "" },
        article: { id: "" },
        user: { username: "" },

        view: "categories",

        printTime: function (item) {
            var t = moment(item);
            t.locale("de");
            return t.fromNow()
        }
    },
    created: function () {
        bus.$on('refresh-categories', this.fetchAllCategories);
        bus.$on('refresh-articles', this.updateArticles);
        bus.$on('refresh-users', this.fetchAllUsers);
        bus.$on('set-token', this.setToken);
    },
    mounted: function () {
        var that = this;
        this.refreshToken(function () {
            that.loadProfile();
            that.fetchAllCategories();
        })
    },
    updated: function () {
        window.componentHandler.upgradeAllRegistered();

        this.$nextTick(function () {
            document.querySelector(".mdl-layout__content").scrollTop = 0;
        });

        $('#mainsearch').unbind('keyup');
        var that = this;
        $("#mainsearch").on('keyup', function (e) {
            if (e.keyCode === 27) {
                $(e.currentTarget).val('');
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
        artLimit: function () { this.fetchAllArticles() },
        artOffset: function () { this.fetchAllArticles() },
    },
    methods: {
        removeToken: function () {
            this.token = "";
            localStorage.removeItem("token");
            location.reload();
        },
        setToken: function (token) {
            this.token = token;
            localStorage.setItem("token", token);
        },
        refreshToken: function (callback) {
            var that = this;
            this.token = localStorage.getItem("token");

            if (this.token) {
                var decoded_token = jwt_decode(this.token);
                if ((Date.now() / 1000) < decoded_token.exp) {
                    $.ajax({
                        url: "/api/authorize/refresh",
                        type: "GET",
                        headers: { "Authorization": "Bearer " + that.token },
                        success: function (data, status, xhr) {
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
            setTimeout(function () { $("body").css("visibility", "visible"); }, 200);
        },
        loadProfile: function () {
            if (this.token) {
                var that = this;
                var decoded_token = jwt_decode(this.token);
                $.ajax({
                    url: "/api/users/" + decoded_token.sub,
                    type: "GET",
                    headers: { "Authorization": "Bearer " + that.token },
                    success: function (data, status, xhr) {
                        that.userProfile = data;
                    }
                });
            }

        },
        incArtOffset: function () {
            this.artOffset += this.artLimit;
        },
        decArtOffset: function () {
            if (this.artOffset >= this.artLimit) {
                this.artOffset -= this.artLimit;
            } else {
                this.artOffset = 0;
            }
        },
        setUserView: function () {
            this.view = "users";
        },
        setArticleView: function () {
            this.view = "articles";
        },
        setCategoryView: function () {
            this.view = "categories";
        },
        fetchAllCategories: function () {
            var that = this;
            var obj = {};
            var list = [];
            $.getJSON("/api/categories", function (json) {
                if (json) {
                    for (var i = 0; i < json.length; i++) {
                        var key = json[i].category;
                        if (!obj[key]) {
                            obj[key] = [];
                        }
                        obj[key].push(json[i]);
                    }

                    var add = function (ob, depth) {
                        for (var i = 0; i < ob.length; i++) {
                            var elem = ob[i];
                            elem.margin = depth * 50;
                            list.push(elem);
                            if (obj[elem.ID]) {
                                add(obj[elem.ID], depth + 1);
                            }
                        }
                    }
                    add(obj[''], 0);

                    that.categories = list;
                } else {
                    that.categories = [];
                }
                that.setCategoryView();
            });
        },
        updateArticles: function (id) {
            var that = this;

            if (!that.articles) {
                that.articles = [];
            }

            $.getJSON('/api/articles/' + id, function (json) {
                for (var i = 0; i < that.articles.length; i++) {
                    if (that.articles[i].ID == id) {
                        that.articles.splice(i, 1, json);
                        return;
                    }
                }
                that.articles.push(json);
            });
        },
        fetchArticles: function (url, callback) {
            var that = this;
            $.getJSON(url, { offset: that.artOffset, limit: that.artLimit }, function (json) {
                that.articles = json;
                that.setArticleView();

                if (callback) {
                    callback();
                }
            });
        },
        fetchArticlesForCategory: function (event) {
            var that = this;
            that.artPaginate = false;
            var id = $(event.target).parents(".mdl-list__item").attr('id');
            this.fetchArticles('/api/articles/category/' + id, function () {
                //that.activeCategory = id;
                that.category = { id: id };
            });
        },
        fetchAllArticles: function () {
            var that = this;
            that.artPaginate = true;
            this.fetchArticles('/api/articles', function () {
                //that.activeCategory = "";
                that.category = { id: "" }
            });
        },
        fetchAllUsers: function () {
            var that = this;
            $.ajax({
                url: "/api/users",
                type: "GET",
                headers: { "Authorization": "Bearer " + that.token },
                success: function (data, status, xhr) {
                    that.users = data;
                    that.setUserView();
                }
            });
        },
        searchCategories: function (query) {
            var that = this;
            $.getJSON("/api/categories/search", { q: query }, function (json) {
                that.categories = json;
                that.setCategoryView();
            })
        },
        searchArticles: function (query) {
            var that = this;
            $.getJSON("/api/articles/search", { q: query }, function (json) {
                that.articles = json;
                that.setArticleView();
            })
        },
        openUserDialog: function (id, method) {
            var that = this;
            this.user.username = id;
            this.$nextTick(function () {
                bus.$emit('showUserDialog', that.token, method);
            });
        },
        openLoginDialog: function (callback) {
            this.$nextTick(function () {
                bus.$emit('showLoginDialog', callback);
            });
        },
        openCategoryDialog: function (id) {
            var that = this;
            this.category.id = id;
            this.$nextTick(function () {
                bus.$emit('showCategoryDialog', that.token);
            });
        },
        openArticleDialog: function (id, category) {
            var that = this;
            this.article.id = id;
            this.$nextTick(function () {
                bus.$emit('showArticleDialog', category, that.token);
            });
        },
        changeUserDialog: function (event) {
            var id = $(event.target).parents("tr").attr('id');
            this.openUserDialog(id, "PUT");
        },
        changeCategoryDialog: function (event) {
            var id = $(event.target).parents(".mdl-list__item").attr('id');
            this.openCategoryDialog(id);
        },
        changeArticleDialog: function (event) {
            var id = $(event.target).parents("tr").attr('id');
            this.openArticleDialog(id);
        },
        newUserDialog: function () {
            this.openUserDialog("", "POST");
        },
        newCategoryDialog: function () {
            this.openCategoryDialog();
        },
        newArticleDialog: function () {
            this.openArticleDialog(null, this.category.id);
        },
        deleteArticle: function (event) {
            var that = this;
            var id = $(event.target).parents("tr").attr('id');
            deleteResource("/api/articles/" + id, this.token, function () {
                for (var i = 0; i < that.articles.length; i++) {
                    if (that.articles[i].ID == id) {
                        that.articles.splice(i, 1);
                        return;
                    }
                }
            })
        },
        deleteCategory: function (event) {
            var that = this;
            var id = $(event.target).parents(".mdl-list__item").attr('id');
            deleteResource("/api/categories/" + id, this.token, that.fetchAllCategories)
        },
        deleteUser: function (event) {
            var that = this;
            var id = $(event.target).parents("tr").attr('id');
            deleteResource("/api/users/" + id, this.token, that.fetchAllUsers)
        },
    }
})

function deleteResource(resource, token, callback) {
    $.ajax({
        url: resource,
        headers: { "Authorization": "Bearer " + token },
        type: "DELETE",
        success: callback
    })
}
