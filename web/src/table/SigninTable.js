// Copyright 2024 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import React from "react";
import {DeleteOutlined, DownOutlined, UpOutlined} from "@ant-design/icons";
import {Button, Col, Input, Popover, Row, Select, Space, Switch, Table, Tooltip} from "antd";
import * as Setting from "../Setting";
import i18next from "i18next";

import {Controlled as CodeMirror} from "react-codemirror2";
import "codemirror/lib/codemirror.css";

require("codemirror/theme/material-darker.css");
require("codemirror/mode/htmlmixed/htmlmixed");

const {Option} = Select;

class SigninTable extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
    };
  }

  updateTable(table) {
    this.props.onUpdateTable(table);
  }

  updateField(table, index, key, value) {
    table[index][key] = value;
    this.updateTable(table);
  }

  addRow(table) {
    const row = {name: Setting.getNewRowNameForTable(table, "Please select a signin item"), visible: true, required: true, rule: "None"};
    if (table === undefined) {
      table = [];
    }
    table = Setting.addRow(table, row);
    this.updateTable(table);
  }

  addCustomRow(table) {
    const randomName = "Text " + Date.now().toString();
    const row = {name: Setting.getNewRowNameForTable(table, randomName), visible: true, isCustom: true};
    if (table === undefined) {
      table = [];
    }
    table = Setting.addRow(table, row);
    this.updateTable(table);
  }

  deleteRow(table, i) {
    table = Setting.deleteRow(table, i);
    this.updateTable(table);
  }

  upRow(table, i) {
    table = Setting.swapRow(table, i - 1, i);
    this.updateTable(table);
  }

  downRow(table, i) {
    table = Setting.swapRow(table, i, i + 1);
    this.updateTable(table);
  }

  renderTable(table) {
    table = table ?? [];
    const columns = [
      {
        title: i18next.t("general:Name"),
        dataIndex: "name",
        key: "name",
        render: (text, record, index) => {
          if (record.isCustom) {
            return <Input style={{width: "100%"}}
              value={text} onPressEnter={e => {
                this.updateField(table, index, "name", e.target.value);
              }} disabled>
            </Input>;
          }

          const items = [
            {name: "Signin methods", displayName: i18next.t("application:Signin methods")},
            {name: "Logo", displayName: i18next.t("general:Logo")},
            {name: "Back button", displayName: i18next.t("login:Back button")},
            {name: "Languages", displayName: i18next.t("general:Languages")},
            {name: "Username", displayName: i18next.t("signup:Username")},
            {name: "Password", displayName: i18next.t("general:Password")},
            {name: "Providers", displayName: i18next.t("general:Providers")},
            {name: "Agreement", displayName: i18next.t("signup:Agreement")},
            {name: "Forgot password?", displayName: i18next.t("login:Forgot password?")},
            {name: "Login button", displayName: i18next.t("login:Login button")},
            {name: "Signup link", displayName: i18next.t("general:Signup link")},
          ];

          const getItemDisplayName = (text) => {
            const item = items.filter(item => item.name === text);
            if (item.length === 0) {
              return "";
            }
            return item[0].displayName;
          };

          return (
            <Select virtual={false} style={{width: "100%"}}
              value={getItemDisplayName(text)}
              onChange={value => {
                this.updateField(table, index, "name", value);
              }} >
              {
                Setting.getDeduplicatedArray(items, table, "name").map((item, index) => <Option key={index} value={item.name}>{item.displayName}</Option>)
              }
            </Select>
          );
        },
      },
      {
        title: i18next.t("organization:Visible"),
        dataIndex: "visible",
        key: "visible",
        width: "120px",
        render: (text, record, index) => {
          if (record.name === "ID") {
            return null;
          }

          return (
            <Switch checked={text} onChange={checked => {
              this.updateField(table, index, "visible", checked);
              if (!checked) {
                this.updateField(table, index, "required", false);
              } else {
                this.updateField(table, index, "required", true);
              }
            }} />
          );
        },
      },
      {
        title: i18next.t("signup:Label HTML"),
        dataIndex: "label",
        key: "label",
        width: "200px",
        render: (text, record, index) => {
          if (record.name.startsWith("Text ") || record?.isCustom) {
            return (
              <Popover placement="right" content={
                <div style={{width: "900px", height: "300px"}} >
                  <CodeMirror value={text}
                    options={{mode: "htmlmixed", theme: "material-darker"}}
                    onBeforeChange={(editor, data, value) => {
                      this.updateField(table, index, "label", value);
                    }}
                  />
                </div>
              } title={i18next.t("signup:Label HTML")} trigger="click">
                <Input value={text} style={{marginBottom: "10px"}} onChange={e => {
                  this.updateField(table, index, "label", e.target.value);
                }} />
              </Popover>
            );
          }
          return null;
        },
      },
      {
        title: i18next.t("application:Form CSS"),
        dataIndex: "label",
        key: "label",
        width: "200px",
        render: (text, record, index) => {
          if (!record.name.startsWith("Text ") && !record?.isCustom) {
            return (
              <Popover placement="right" content={
                <div style={{width: "900px", height: "300px"}} >
                  <CodeMirror value={text}
                    options={{mode: "htmlmixed", theme: "material-darker"}}
                    onBeforeChange={(editor, data, value) => {
                      this.updateField(table, index, "label", value);
                    }}
                  />
                </div>
              } title={i18next.t("application:CSS style")} trigger="click">
                <Input value={text} onChange={e => {
                  this.updateField(table, index, "label", e.target.value);
                }} />
              </Popover>
            );
          }

          return null;
        },
      },
      {
        title: i18next.t("signup:Placeholder"),
        dataIndex: "placeholder",
        key: "placeholder",
        width: "200px",
        render: (text, record, index) => {
          if (record.name !== "Username" && record.name !== "Password") {
            return null;
          }

          return (
            <Input value={text} onChange={e => {
              this.updateField(table, index, "placeholder", e.target.value);
            }} />
          );
        },
      },
      {
        title: i18next.t("application:Rule"),
        dataIndex: "rule",
        key: "rule",
        width: "155px",
        render: (text, record, index) => {
          let options = [];
          if (record.name === "Providers") {
            options = [
              {id: "big", name: i18next.t("application:Big icon")},
              {id: "small", name: i18next.t("application:Small icon")},
            ];
          }
          if (options.length === 0) {
            return null;
          }

          return (
            <Select virtual={false} style={{width: "100%"}} value={text} onChange={(value => {
              this.updateField(table, index, "rule", value);
            })} options={options.map(item => Setting.getOption(item.name, item.id))} />
          );
        },
      },
      {
        title: i18next.t("general:Action"),
        key: "action",
        width: "100px",
        render: (text, record, index) => {
          return (
            <div>
              <Tooltip placement="bottomLeft" title={i18next.t("general:Up")}>
                <Button style={{marginRight: "5px"}} disabled={index === 0} icon={<UpOutlined />} size="small" onClick={() => this.upRow(table, index)} />
              </Tooltip>
              <Tooltip placement="topLeft" title={i18next.t("general:Down")}>
                <Button style={{marginRight: "5px"}} disabled={index === table.length - 1} icon={<DownOutlined />} size="small" onClick={() => this.downRow(table, index)} />
              </Tooltip>
              <Tooltip placement="topLeft" title={i18next.t("general:Delete")}>
                <Button icon={<DeleteOutlined />} size="small" onClick={() => this.deleteRow(table, index)} />
              </Tooltip>
            </div>
          );
        },
      },
    ];

    return (
      <Table scroll={{x: "max-content"}} rowKey="name" columns={columns} dataSource={table} size="middle" bordered pagination={false}
        title={() => (
          <Space>
            {this.props.title}
            <Button style={{marginRight: "5px"}} type="primary" size="small" onClick={() => this.addRow(table)}>{i18next.t("general:Add")}</Button>
            <Button style={{marginRight: "5px"}} type="primary" size="small" onClick={() => this.addCustomRow(table)}>{i18next.t("general:Add custom item")}</Button>
          </Space>
        )}
      />
    );
  }

  render() {
    return (
      <div>
        <Row style={{marginTop: "20px"}} >
          <Col span={24}>
            {
              this.renderTable(this.props.table)
            }
          </Col>
        </Row>
      </div>
    );
  }
}

export default SigninTable;
