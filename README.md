# Where I am at

* push works, but slides don't update on ud backend.  need to change it so it respects `params[:deck][:slides]`.

---

# Where I am at - Aug 24 2017

* push now works.
* should now be easy to implement pull and watch, now that we're centering around pushing + pulling .ud.json.
* [ ] the frontend is probably broken, because I changed the api sig on the backend to support the cli.
* upgrade account works well (maybe rename "upgrade" to something else.)

# TODO next

* [x] Implement pull
* [x] Implement updated_at timestamp checking for push + pull
* [ ] Force command to force-push or force-pull
* [ ] Implement watch
* [ ] asset push / sync
* [ ] webhook support on frontend to support auto-updating
* [ ] Fix frontend (if needed b/c of potential breaking api changes)

# TODO once almost done

* [ ] homebrew
* [ ] other (legit) sites that host binaries

---

If pushing:

* updated_at Timestamp on client must be equal to or greater than updated_at timestamp on server
* will need to do a GET request check for that

If pulling

* updated_at on SERVER must be equal to or greater than updated_at timestamp on server
* updated_at on
