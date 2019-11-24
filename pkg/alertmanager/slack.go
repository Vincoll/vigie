package alertmanager

type slackAlert struct {
	webhookURL string
}

/*
//https://api.slack.com/tools/block-kit-builder?mode=message&blocks=%5B%7B%22type%22%3A%22section%22%2C%22text%22%3A%7B%22type%22%3A%22mrkdwn%22%2C%22text%22%3A%22Hello%2C%20Assistant%20to%20the%20Regional%20Manager%20Dwight!%20*Michael%20Scott*%20wants%20to%20know%20where%20you%27d%20like%20to%20take%20the%20Paper%20Company%20investors%20to%20dinner%20tonight.%5Cn%5Cn%20*Please%20select%20a%20restaurant%3A*%22%7D%7D%2C%7B%22type%22%3A%22divider%22%7D%2C%7B%22type%22%3A%22section%22%2C%22text%22%3A%7B%22type%22%3A%22mrkdwn%22%2C%22text%22%3A%22*Ler%20Ros*%5Cn%3Astar%3A%3Astar%3A%3Astar%3A%3Astar%3A%202082%20reviews%5Cn%20I%20would%20really%20recommend%20the%20%20Yum%20Koh%20Moo%20Yang%20-%20Spicy%20lime%20dressing%20and%20roasted%20quick%20marinated%20pork%20shoulder%2C%20basil%20leaves%2C%20chili%20%26%20rice%20powder.%22%7D%7D%2C%7B%22type%22%3A%22section%22%2C%22text%22%3A%7B%22type%22%3A%22mrkdwn%22%2C%22text%22%3A%22*Ler%20Ros*%5Cn%3Astar%3A%3Astar%3A%3Astar%3A%3Astar%3A%202082%20reviews%5Cn%20I%20would%20really%20recommend%20the%20%20Yum%20Koh%20Moo%20Yang%20-%20Spicy%20lime%20dressing%20and%20roasted%20quick%20marinated%20pork%20shoulder%2C%20basil%20leaves%2C%20chili%20%26%20rice%20powder.%22%7D%7D%2C%7B%22type%22%3A%22divider%22%7D%2C%7B%22type%22%3A%22section%22%2C%22fields%22%3A%5B%7B%22type%22%3A%22plain_text%22%2C%22text%22%3A%22*this%20is%20plain_text%20text*%22%2C%22emoji%22%3Atrue%7D%2C%7B%22type%22%3A%22plain_text%22%2C%22text%22%3A%22*this%20is%20plain_text%20text*%22%2C%22emoji%22%3Atrue%7D%2C%7B%22type%22%3A%22plain_text%22%2C%22text%22%3A%22*this%20is%20plain_text%20text*%22%2C%22emoji%22%3Atrue%7D%2C%7B%22type%22%3A%22plain_text%22%2C%22text%22%3A%22*this%20is%20plain_text%20text*%22%2C%22emoji%22%3Atrue%7D%2C%7B%22type%22%3A%22plain_text%22%2C%22text%22%3A%22*this%20is%20plain_text%20text*%22%2C%22emoji%22%3Atrue%7D%5D%7D%5D
var ex = `
{
"blocks": [
{
"type": "section",
"text": {
"type": "mrkdwn",
"text": "Hello, Assistant to the Regional Manager Dwight! *Michael Scott* wants to know where you'd like to take the Paper Company investors to dinner tonight.\n\n *Please select a restaurant:*"
}
},
{
"type": "divider"
},
{
"type": "section",
"text": {
"type": "mrkdwn",
"text": "*Ler Ros*\n:star::star::star::star: 2082 reviews\n I would really recommend the  Yum Koh Moo Yang - Spicy lime dressing and roasted quick marinated pork shoulder, basil leaves, chili & rice powder."
}

},
{
"type": "section",
"text": {
"type": "mrkdwn",
"text": "*Ler Ros*\n:star::star::star::star: 2082 reviews\n I would really recommend the  Yum Koh Moo Yang - Spicy lime dressing and roasted quick marinated pork shoulder, basil leaves, chili & rice powder."
}

},
{
"type": "divider"
}
]
}
`


*/
