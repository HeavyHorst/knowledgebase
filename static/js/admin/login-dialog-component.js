Vue.component("login-dialog", {
  data: function() {
    return {
      token: "",
      username: "",
      password: "",
      afterLogin: function() {}
    };
  },
  created: function() {
    bus.$on("showLoginDialog", this.showLoginDialog);
  },
  destroyed: function() {
    bus.$off("showLoginDialog", this.showLoginDialog);
  },
  mounted: function() {
    var that = this;
    $(this.$refs.loginDialog.querySelectorAll("input")).on("keydown", function(
      e
    ) {
      // Enter pressed?
      if (e.which == 10 || e.which == 13) {
        that.authenticate();
      }
    });
  },
  template: ' \
    <dialog class="mdl-dialog login-dialog" ref="loginDialog" style="width:400px !important"> \
    <h4 class="mdl-dialog__title">LOG IN</h4> \
    <hr> \
    <div class="mdl-dialog__content"> \
      <form action="" class="cf"> \
        <div class="mdl-selectfield cf"> \
          <div class="cf"> \
            <input v-model="username" type="text" placeholder="Username"> \
          </div> \
          <div class="cf"> \
            <input v-model="password" type="password" placeholder="Password"> \
          </div> \
        </div> \
      </form> \
    </div> \
    <div class="mdl-dialog__actions"> \
      <button type="submit" class="mdl-button mdl-js-button mdl-button--raised mdl-js-ripple-effect" v-on:click="authenticate">Login</button> \
    </div> \
  </dialog> \
    ',
  methods: {
    showLoginDialog: function(callback) {
      var that = this;
      this.afterLogin = callback;
      var dialog = this.$refs.loginDialog;

      dialog.addEventListener(
        "close",
        function(event) {
          if (!that.token) {
            dialog.showModal();
          }
        },
        false
      );

      // TODO - warum verschwindet der backdrop sofort ohne den Timeout ?
      setTimeout(function() {
        dialog.showModal();
      }, 100);
    },
    authenticate: function() {
      var that = this;
      var dialog = this.$refs.loginDialog;

      $.ajax({
        url: "/api/authorize",
        type: "POST",
        data: {
          password: that.password,
          username: that.username
        },
        success: function(data, status, xhr) {
          that.token = data.token;
          bus.$emit("set-token", data.token);
          dialog.close();
          that.afterLogin();
        }
      });
    }
  }
});
