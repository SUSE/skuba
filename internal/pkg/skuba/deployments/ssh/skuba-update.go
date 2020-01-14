/*
 * Copyright (c) 2019 SUSE LLC.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package ssh

func init() {
	stateMap["skuba-update.start.no-block"] = skubaUpdateStartNoWait
	stateMap["skuba-update-timer.enable"] = skubaUpdateTimerEnable
	stateMap["skuba-update-timer.disable"] = skubaUpdateTimerDisable
}

func skubaUpdateStartNoWait(t *Target, data interface{}) error {
	_, _, err := t.ssh("systemctl", "start", "--no-block", "skuba-update")
	return err
}

func skubaUpdateTimerEnable(t *Target, data interface{}) error {
	_, _, err := t.ssh("systemctl", "enable", "--now", "skuba-update.timer")
	return err
}

func skubaUpdateTimerDisable(t *Target, data interface{}) error {
	_, _, err := t.ssh("systemctl", "disable", "--now", "skuba-update.timer")
	return err
}
