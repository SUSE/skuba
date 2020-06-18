class PrStatus:
    def __init__(self, repo, change_id, build_url):
        self.repo = repo
        self.change_id = change_id
        self.build_url = build_url

    def _create_pr_status(self, context, description, state):
        pr = self.repo.get_pull(self.change_id)
        commit = self.repo.get_commit(sha=pr.head.sha)
        commit.create_status(state=state,
                             target_url=f'{self.build_url}display/redirect',
                             description=description,
                             context=context)

    def update_pr_status(self, context, state):
        if state == 'error':
            self._create_pr_status(context, 'error', state)
        elif state == 'failure':
            self._create_pr_status(context, 'failed', state)
        elif state == 'pending':
            self._create_pr_status(context, 'in-progress', state)
        elif state == 'success':
            self._create_pr_status(context, 'success', state)
        else:
            raise Exception('Unknown state')
