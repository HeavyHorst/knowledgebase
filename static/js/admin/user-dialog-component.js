Vue.component("user-dialog", {
    props: ["username"],
    data: function () {
        return {
            user: {
                username: "",
                first_name: "",
                last_name: "",
                is_admin: false,
                password: "",
                image: "",
            },
            token: "",
            method: "POST",
        }
    },
    created: function () {
        bus.$on('showUserDialog', this.showUserDialog);
    },
    destroyed: function () {
        bus.$off('showUserDialog', this.showUserDialog);
    },
    template: ' \
    <dialog class="mdl-dialog animate" ref="userDialog" style="width:30% !important"> \
    <h4 class="mdl-dialog__title">Benutzer:</h4> \
    <hr> \
    <div class="mdl-dialog__content"> \
      <form action="" class="cf"> \
        <div class="cf"> \
          <div class="cf half left"> \
            <input ref="usernameInput" v-model="user.username" type="text" placeholder="Username"> \
          </div> \
          <div class="cf half right"> \
            <input v-model="user.password" type="password" placeholder="Password"> \
          </div> \
          <div class="cf half left"> \
            <input v-model="user.first_name" type="text" placeholder="Vorname"> \
          </div> \
          <div class="cf half right"> \
            <input v-model="user.last_name" type="text" placeholder="Nachname"> \
          </div> \
          <div class="cf"> \
            <input v-model="user.image" type="text" placeholder="Bild"> \
          </div> \
          <div class="cf half left" style="margin-top: 1em;"> \
              <span> Admin: </span> \
          </div> \
          <div class="cf half right" style="margin-top: 1em;"> \
              <input type="checkbox" id="checkbox" v-model="user.is_admin"> \
          </div> \
        </div> \
      </form> \
    </div> \
    <div class="mdl-dialog__actions"> \
      <button type="button" class="mdl-button mdl-js-button mdl-button--raised mdl-js-ripple-effect" v-on:click="updateUser">Speichern</button> \
    </div> \
  </dialog> \
    ',
    methods: {
        getUser: function (callback) {
            if (this.username) {
                var that = this;
                $.ajax({
                    url: "/api/users/" + that.username,
                    type: "GET",
                    headers: { "Authorization": "Bearer " + that.token },
                    success: function (data, status, xhr) {
                        that.user = data;
                        that.$refs.usernameInput.disabled = true;
                    }
                });
            } else {
                this.user = {};
                this.$refs.usernameInput.disabled = false;
            }
        },
        showUserDialog: function (token, method) {
            this.token = token;

            if (method) {
                this.method = method;
            }

            var dialog = this.$refs.userDialog;
            this.getUser();
            dialog.showModal();
        },
        updateUser: function () {
            var that = this;
            var dialog = this.$refs.userDialog;

            if (this.$refs.usernameInput.value.trim() === "") {
                alert("Sie müssen einen Benutzernamen angeben!");
                return;
            }

            var url = "/api/users/" + this.user.username;
            if (this.method === "POST") {
                var url = '/api/users';
            }

            $.ajax({
                url: url,
                type: that.method,
                headers: { "Authorization": "Bearer " + that.token },
                data: JSON.stringify(that.user),
                contentType: "application/json; charset=utf-8",
                success: function () {
                    bus.$emit('refresh-users');
                    dialog.close();
                }
            });
        },
    }
})