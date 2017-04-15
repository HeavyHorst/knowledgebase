Vue.component("category-dialog", {
  props: ["cid"],
  data: function() {
    return {
      category: {
        id: "",
        category: "",
        title: "",
        image: "",
        description: ""
      },
      categories: [],
      token: ""
    };
  },
  created: function() {
    this.$nextTick(function() {
      this.simplemde = new SimpleMDE({
        element: this.$refs.descedit,
        spellChecker: false,
        autoDownloadFontAwesome: false,
        forceSync: true
      });
    });

    bus.$on("showCategoryDialog", this.showCategoryDialog);
  },
  destroyed: function() {
    bus.$off("showCategoryDialog", this.showCategoryDialog);
  },
  template: '\
<dialog class="mdl-dialog animate" ref="categoryDialog"> \
    <h4 class="mdl-dialog__title">Kategorie:</h4> \
    <hr> \
    <div class="mdl-dialog__content"> \
      <form action=""> \
        <input v-model="category.id" type="hidden" name="ID"> \
        <div class="cf half left"> \
          <input v-model="category.title" type="text" placeholder="Titel"> \
        </div> \
        <div class="mdl-selectfield cf half right"> \
          <select v-model="category.category"> \
            <option value="">Keine Kategorie</option> \
            <option v-for="cat in categories" :value="cat.ID"> {{ cat.title }} </option> \
          </select> \
        </div> \
        <div class="cf"> \
          <input v-model="category.image" type="text" placeholder="Bild"> \
        </div> \
        <div class="mdl-textfield mdl-js-textfield"> \
          <textarea v-model="category.description" ref="descedit" class="mdl-textfield__input" type="text" rows= "10"></textarea> \
        </div> \
      </form> \
    </div> \
    <div class="mdl-dialog__actions"> \
      <button type="button" class="mdl-button mdl-js-button mdl-button--raised mdl-js-ripple-effect" v-on:click="updateCategory" >Speichern</button> \
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
    setDescription: function(text) {
      this.simplemde.value(text);
      var that = this;
      setTimeout(
        function() {
          that.simplemde.codemirror.refresh();
        }.bind(this.simplemde),
        0
      );
    },
    getCategory: function() {
      var that = this;
      if (this.cid) {
        $.ajax({
          url: "/api/categories/" + that.cid,
          type: "GET",
          headers: { Authorization: "Bearer " + that.token },
          success: function(cat) {
            that.getCategories(function() {
              that.category = cat;
              that.setDescription(cat.description);
            });
          }
        });
      } else {
        that.getCategories(function() {
          that.category = { category: "" };
          that.setDescription("");
        });
      }
    },
    showCategoryDialog: function(token) {
      this.token = token;
      var dialog = this.$refs.categoryDialog;
      this.getCategory();
      dialog.showModal();
    },
    updateCategory: function() {
      var that = this;
      var dialog = this.$refs.categoryDialog;
      var method = "PUT";
      var url = "/api/categories/" + this.category.ID;
      if (!this.category.ID) {
        method = "POST";
        var url = "/api/categories";
      }

      // warum geht forceSync nicht ?
      this.category.description = this.simplemde.value();

      $.ajax({
        url: url,
        type: method,
        headers: { Authorization: "Bearer " + that.token },
        data: JSON.stringify(that.category),
        contentType: "application/json; charset=utf-8",
        success: function() {
          bus.$emit("refresh-categories");
          dialog.close();
        }
      });
    }
  }
});
