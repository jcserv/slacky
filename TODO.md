# current work
main view: 
 - statusbar should show the current selected channel
 - should also load in DMs

# mvp

core functionality
  - channels
    - top bar (links, folders) 
    - star channel
  - dms
  - command bar
    - # --> query channels
    - @ --> query people
  - search
    - sort (most relevant / newest / oldest) 
    - filters
      - location (messages/dms/files/people/channels/canvases)
      - from (@people)
      - in: #channel / @people / apps
      - only my channels: true/false
      - exclude automations
      - from: @people
      - with: @people
      - date
        - any time / today / yesterday / last 7/30 days / last 3 months / last 12 months
        - on
        - before
        - after
        - range
      - file types (lists / canvases & posts / documents / emails / images / pdfs / presentations / snippets / spreadsheets / audio / videos )
      - reactions
        - from you
        - from anyone
      - message has (file / action / link)
      - message is (dm / thread / saved / pinned)  
  - activity -> recent messages
  - file attachments

user
  - set status
  - reply in thread

message
  - send message
  - schedule message
  - react; reactions
  - delete
  - save message for later
  - mark as unread
  - pin to channel
  - copy link
  - forward message
  - remind me

message input
    - @ dropdown
    - attach file

    text formatting
    - bold (ctrl+b)
    - italics (ctrl+i)
    - strikethrough (ctrl+shift+x)
    - link (ctrl+shift+u)
    - ordered list (ctrl+shift+7)
    - bulleted list (ctrl+shift+8)
    - blockquote (ctrl+shift+9)
    - code (ctrl+shift+c)
    - code block (ctrl+option+shift+c)

ui/ux:
- keybind to copy current channel
  - name
  - link
  - huddle link 

add channels/conversations to tabs

# eventually
- channel management
  - add link to top bar
  - add folder to top bar
- directories
- multi-workspace support
- history
  - prev/forward 

# maybe
- summarize channels/threads
  - implement it on our end? since their summarize thread feature is not super helpful from my experience
- files tab
- canvases
- text snippet

# out of scope
- huddles
- videos
- audio clips
- workflows