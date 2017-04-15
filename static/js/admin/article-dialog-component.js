Vue.component("article-dialog", {
  props: ["aid"],
  data: function() {
    return {
      article: {
        id: "",
        title: "",
        category: "",
        short: "",
        article: "",
        tags: ""
      },
      categories: [],
      token: ""
    };
  },
  created: function() {
    this.$nextTick(function() {
      this.simplemde = new SimpleMDE({
        element: this.$refs.artedit,
        spellChecker: false,
        autoDownloadFontAwesome: false,
        forceSync: true
      });
    });

    bus.$on("showArticleDialog", this.showArticleDialog);
  },
  destroyed: function() {
    bus.$off("showArticleDialog", this.showArticleDialog);
  },
  template: ' \
    <dialog class="mdl-dialog animate" ref="articleDialog"> \
    <h4 class="mdl-dialog__title">Artikel:</h4> \
    <hr> \
    <div class="mdl-dialog__content"> \
      <form action="" class="cf"> \
        <input v-model="article.id" type="hidden" name="ID" value=""> \
        <div class="mdl-selectfield cf half left"> \
          <input v-model="article.title" type="text" placeholder="Titel"> \
          <select v-model="article.category"> \
            <option value="" disabled>Wähle eine Kategorie aus</option> \
            <option v-for="cat in categories" :value="cat.ID"> {{ cat.title }} </option> \
          </select> \
        </div> \
        <div class="cf half right"> \
          <textarea v-model="article.short" type="text" placeholder="Beschreibung"></textarea> \
        </div> \
        <div class="mdl-textfield mdl-js-textfield cf"> \
          <input class="mdl-textfield__input" type="text" ref="dialog_article_tags"> \
        </div> \
        <div class="mdl-textfield mdl-js-textfield"> \
          <textarea v-model="article.article" class="mdl-textfield__input" type="text" rows="10" ref="artedit" ></textarea> \
        </div> \
      </form> \
    </div> \
    <div class="mdl-dialog__actions"> \
      <button type="button" class="mdl-button mdl-js-button mdl-button--raised mdl-js-ripple-effect" v-on:click="updateArticle">Speichern</button> \
    </div> \
  </dialog> \
    ',
  methods: {
    getCategories: function(callback) {
      var that = this;
      $.ajax({
        url: "/api/categories",
        type: "GET",
        headers: { Authorization: "Bearer " + that.token },
        success: function(json) {
          that.categories = json;

          if (callback) {
            callback();
          }
        }
      });
    },
    setArticleText: function(text) {
      this.simplemde.value(text);
      var that = this;
      setTimeout(
        function() {
          that.simplemde.codemirror.refresh();
        }.bind(this.simplemde),
        0
      );
    },
    getArticle: function(category) {
      if (this.aid) {
        var that = this;
        $.ajax({
          url: "/api/articles/" + that.aid,
          type: "GET",
          headers: { Authorization: "Bearer " + that.token },
          success: function(art) {
            if (!art.tags) {
              art.tags = [];
            }

            that.getCategories(function() {
              $(that.$refs.dialog_article_tags).importTags(art.tags.join());
              $(that.$refs.dialog_article_tags).tagsInput({ width: "100%" });

              that.article = art;
              that.setArticleText(art.article);
            });
          }
        });
      } else {
        var that = this;
        this.getCategories(function() {
          that.article = { category: category };
          $(that.$refs.dialog_article_tags).importTags("");
          $(that.$refs.dialog_article_tags).tagsInput({ width: "100%" });
          that.setArticleText("");
        });
      }
    },
    showArticleDialog: function(category, token) {
      this.token = token;
      var dialog = this.$refs.articleDialog;
      this.getArticle(category);
      dialog.showModal();
    },
    updateArticle: function() {
      var that = this;
      var dialog = this.$refs.articleDialog;
      var method = "PUT";
      var url = "/api/articles/" + this.article.ID;
      if (!this.article.ID) {
        method = "POST";
        var url = "/api/articles";
      }

      // warum geht forceSync nicht ?
      this.article.article = this.simplemde.value();
      this.article.tags = this.$refs.dialog_article_tags.value.split(",");

      $.ajax({
        url: url,
        type: method,
        headers: { Authorization: "Bearer " + that.token },
        data: JSON.stringify(that.article),
        contentType: "application/json; charset=utf-8",
        success: function(data, status, xhr) {
          var location = xhr.getResponseHeader("Location");
          var id = that.article.ID;
          if (location) {
            var p = location.split("/");
            id = p[p.length - 1];
          }
          bus.$emit("refresh-articles", id);
          dialog.close();
        }
      });
    }
  }
});
