Gmail Unread-Read
=================

If you're anything like me, your personal email inbox has an excessive 
amount of unread emails. Enough that it would be impossible to actually 
go through and mark them all as read. 

That's why I made this project.

This project uses the Gmail API via GCP to mark ***ALL*** unread messages
in your inbox as read. Be warned though, this can take a very long time.

For me, approx. 33k emails took around 2 hours.

## Usage / Setup

To use this you must have Golang installed on your system. Then use the 
`make` command to build and / or run the `unreadread` command.

The command does not take in any command line arguments or read any 
environmental variables. It only reads the `credentials.json` file 
file from the directory it is executed from.

The `credentials.json` file has to come from the Google Cloud Platform (GCP).
To obtain this you have to create a project which has access to the Gmail API. 

Below are a rough set of instructions to set this up:
1. Go to the [GCP Console](https://console.cloud.google.com/)
2. Create a New Project (call it whatever you want) and select the new project
3. In the right menu, go to `APIs & Services > Enable APIs & Services`
4. Click the `+ ENABLE APIS AND SERVICES` button
5. Search "Gmail API", select it, and hit the "Enable" button
    > [This may be the link for it.](https://console.cloud.google.com/apis/library/gmail.googleapis.com)
6. After enabling it, hit the "CREATE CREDENTIALS" button
    - Select the "GMail API"
    - Select "User data"
7. In the "Scopes" options, add the following
    - `https://mail.google.com/`
    - `https://www.googleapis.com/auth/gmail.modify`
8. For the "OAuth Client ID" Application Type, select "Desktop app"
9. Download the credentials at the end, save them to the `credentials.json` file in this application
10. In "Audience" for the app, add your email you wish to clear of unread emails to the "Test users"

You should ideally be good to go.

## Known Issues

### Continual Enqueuing of "unread" Emails

The way this application works is it queries the Gmail API to give it
unread emails, which comes back in batches of 100. For some reason, the 
`NextPageToken` token might not be being used correctly.

The fix for this is to re-run this as you get to the end of the unread emails.

## Useful Resources
