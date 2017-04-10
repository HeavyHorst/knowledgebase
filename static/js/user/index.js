$(document).ready(function () {
    // search event handler
    $('#mainsearch').unbind('keyup');
    $("#mainsearch").on('keyup', function (e) {
        if (e.keyCode === 27) {
            $(e.currentTarget).val('');
            app.fetchAllCategories();
        } else if (e.keyCode === 13) {
            app.searchArticles(e.currentTarget.value);
        }
    });

    var bus = new Vue();
    app = new Vue({
        el: '#content',
        data: function () {
            return {
                view: "categories",
                categories: [],
                subCategories: [],
                articles: [],
                article: "",
                url_path: "/categories",
                theme: "",
                printTime: function (item) {
                    var t = moment(item);
                    t.locale("de");
                    return t.fromNow()
                }
            }
        },
        created: function () {
            this.token = localStorage.getItem("token");

            window.onpopstate = (event) => {
                if (event.state === null) {
                    if (window.location.hash) {
                        document.querySelector(window.location.hash).scrollIntoView();
                    }
                    event.preventDefault();
                    return false;
                }
                this.parseRoute();
            };

            window.addEventListener('resize', this.scaleGrid)
        },
        mounted: function () {
            var that = this;
            this.theme = localStorage.getItem("theme");

            var theme_loaded = function () {
                that.parseRoute();
                bus.$off('theme-loaded', theme_loaded);
            }

            bus.$on('theme-loaded', theme_loaded);
        },
        updated: function () {
            var that = this;
            this.$nextTick(function () {
                if (this.view === "categories") {
                    $("#content").css("width", "1000px");
                } else if (this.view === "articles") {
                    $("#content").css("width", "95%");
                    that.$nextTick(function () {
                        this.scaleGrid();
                    });
                } else if (this.view === "article") {
                    $("#content").css("width", "100%");
                }

                if (!window.location.hash) {
                    document.querySelector(".mdl-layout__content").scrollTop = 0;
                }
            });
        },
        watch: {
            url_path: function (val) {
                history.pushState({}, "", this.url_path);
            },
            theme: function (val) {
                if (val === "dark") {
                    this.intoDarkness();
                } else {
                    this.intoTheLight();
                }
            }
        },
        methods: {
            scaleGrid: function () {
                var cwidth = document.querySelector("#content").clientWidth;
                var gewidth = 0;
                var num = 8;

                while (gewidth < 500) {
                    num -= 1;
                    gewidth = (cwidth - num * 30) / num;
                }

                if (cwidth < 800) {
                    $(".demo-card-square").css("width", "100%");
                } else {
                    $(".demo-card-square").css("width", gewidth);
                }
            },
            setTheme: function (theme) {
                this.theme = theme;
            },
            intoDarkness: function () {
                $('head').append($('<link rel="stylesheet" type="text/css" />').attr('href', '/static/css/dark.css'));
                $('head').append($('<link rel="stylesheet" type="text/css" />').attr('href', '/static/css/highlightjs/gruvbox-dark.css'));
                $("#masterofdarkness").removeClass("menu__item--current");
                $("#bringeroflight").addClass("menu__item--current");
                localStorage.setItem("theme", "dark");

                // a little hacky - wait for the css to load
                var fakeListener = setInterval(function () {
                    console.log($(".mdl-layout__content").css("background"))
                    if ($(".mdl-layout__content").css("background") === "rgb(0, 0, 0) none repeat scroll 0% 0% / auto padding-box border-box") {
                        clearInterval(fakeListener)
                        bus.$emit('theme-loaded');
                    }
                }, 50)
            },
            intoTheLight: function () {
                $("LINK[href='/static/css/dark.css']").remove();
                $("LINK[href='/static/css/highlightjs/gruvbox-dark.css']").remove();
                $("#masterofdarkness").addClass("menu__item--current");
                $("#bringeroflight").removeClass("menu__item--current");
                localStorage.setItem("theme", "light");
                bus.$emit('theme-loaded');
            },
            fetchAllCategories: function () {
                var that = this;
                $.ajax({
                    url: "/api/categories",
                    type: "GET",
                    headers: { "Authorization": "Bearer " + that.token },
                    success: function (json) {
                        that.categories = json;
                        that.view = "categories";
                    }
                });

                this.url_path = "/categories";
            },
            fetchCategoryChilds: function (event) {
                var that = this;
                var id = null;

                if (typeof (event) === "string") {
                    id = event;
                } else {
                    id = $(event.target).closest("table").attr("id");
                }

                var categories, articles
                $.when(
                    $.ajax({
                        url: '/api/categories/category/' + id,
                        type: "GET",
                        headers: { "Authorization": "Bearer " + that.token },
                        success: function (json) {
                            categories = json;
                        }
                    }),
                    $.ajax({
                        url: '/api/articles/category/' + id,
                        type: "GET",
                        headers: { "Authorization": "Bearer " + that.token },
                        success: function (json) {
                            articles = json;
                        }
                    })
                ).then(function () {
                    that.subCategories = categories;
                    that.articles = articles;
                    that.view = "articles";
                });

                this.url_path = "/articles/category/" + id;
            },
            showArticle: function (event) {
                var that = this;
                var id = null;

                if (typeof (event) === "string") {
                    id = event;
                } else {
                    id = $(event.target).closest(".demo-card-square").attr("id");
                }

                $.ajax({
                    url: '/api/articles/' + id,
                    type: "GET",
                    headers: { "Authorization": "Bearer " + that.token },
                    success: function (json) {
                        var converter = new showdown.Converter();
                        converter.setFlavor('github');
                        that.article = converter.makeHtml("# " + json.title + "\n\n" + json.article);

                        that.view = "article";

                        // generate toc and scroll to anker element
                        that.$nextTick(function () {
                            $('h1,h2,h3,h4').each(function (i, val) {
                                if (!val.id) {
                                    val.id = val.innerText.replace(/\s/g, '').toLowerCase();
                                }
                                val.innerHTML = '<a href="#' + val.id + '">' + val.innerText + '</a>';
                            });

                            $('pre code').each(function (i, block) {
                                hljs.highlightBlock(block);
                            });

                            if (window.location.hash) {
                                document.querySelector(window.location.hash).scrollIntoView();
                            }
                        })
                    }
                });

                this.url_path = "/articles/" + id + window.location.hash;
            },
            searchArticles: function (query) {
                that = this;
                $.ajax({
                    url: '/api/articles/search',
                    type: "GET",
                    headers: { "Authorization": "Bearer " + that.token },
                    data: { q: query },
                    success: function (json) {

                        if (json.length === 1) {
                            that.showArticle(json[0].ID);
                            return
                        }

                        that.subCategories = [];
                        that.articles = json;
                        that.view = "articles";
                    }
                });

                this.url_path = "/articles/search?" + $.param({ q: query });
            },
            parseRoute: function () {
                // very simple url routing
                var that = this;
                var path = window.location.href;

                var routes = [
                    ["/categories", this.fetchAllCategories],
                    ["/articles/category/:catid", this.fetchCategoryChilds],
                    ["/articles/search", function () {
                        var query = deparam("q=" + path.split("?q=")[1]);
                        that.searchArticles(query.q)
                    }],
                    ["/articles/:artid", this.showArticle],
                    ["/", this.fetchAllCategories],
                ]

                for (var i = 0; i < routes.length; i++) {
                    var route = routes[i][0];
                    var routeMatcher = new RegExp(route.replace(/:[^\s/]+/g, '([\\w-]+)'));
                    var match = path.match(routeMatcher);
                    if (match) {
                        routes[i][1](match[1]);
                        break;
                    }
                }
            },
        }
    })
});