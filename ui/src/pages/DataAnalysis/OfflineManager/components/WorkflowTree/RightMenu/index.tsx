import { Menu, message } from "antd";
import {
  OfflineRightMenuClickSourceEnums,
  PrimaryEnums,
  SecondaryEnums,
  TertiaryEnums,
} from "@/pages/DataAnalysis/service/enums";
import { useCallback, useEffect, useMemo, useRef } from "react";
import { ItemType } from "antd/es/menu/hooks/useItems";
import { AppstoreAddOutlined, EditOutlined } from "@ant-design/icons";
import IconFont from "@/components/IconFont";
import { useModel } from "@@/plugin-model/useModel";
import { useIntl } from "umi";
import deletedModal from "@/components/DeletedModal";
import lodash from "lodash";
// import SVGIcon, { SVGTypeEnums } from "@/components/SVGIcon";
import useLocalStorages, { LocalModuleType } from "@/hooks/useLocalStorages";
import useUrlState from "@ahooksjs/use-url-state";

export interface RightMenuProps {
  clickSource: OfflineRightMenuClickSourceEnums;
  currentNode?: any;
  handleCloseNodeModal?: (params?: any) => void;
}
const RightMenu = (props: RightMenuProps) => {
  const i18n = useIntl();
  const { clickSource, currentNode, handleCloseNodeModal } = props;
  const { onSetLocalData } = useLocalStorages();
  const [_, setUrlState] = useUrlState();
  const selectNodeRef = useRef<any>(null);
  const { workflow, currentInstances, manageNode } = useModel("dataAnalysis");
  const {
    setVisibleWorkflowEditModal,
    setEditWorkFlow,
    setIsEditWorkflow,
    getWorkflow,
    deleteWorkflow,
    getWorkflows,
    setWorkflowList,
  } = workflow;

  const {
    showNodeModal,
    showFolderModal,
    setExtra,
    setIsEditNode,
    setCurrentNode,
    doDeletedNode,
    doDeleteFolder,
    setSelectNode,
  } = manageNode;

  const handleClickAddWorkflow = useCallback(
    () => setVisibleWorkflowEditModal(true),
    []
  );

  // const handleClickBoard = useCallback(() => {
  //   setSelectNode(currentNode.board);
  //   selectNodeRef.current = currentNode.board;
  // }, [currentNode]);

  useEffect(() => {
    currentNode?.board && (selectNodeRef.current = currentNode.board);
  }, [currentNode, currentNode?.board]);

  const handleClickUpdateWorkflow = useCallback(() => {
    if (!currentNode) return;
    getWorkflow.run(currentNode.id).then((res) => {
      if (res?.code !== 0) return;
      setVisibleWorkflowEditModal(() => {
        setEditWorkFlow(res.data);
        setIsEditWorkflow(true);
        return true;
      });
    });
  }, [currentNode]);

  const handleClickDeleteWorkflow = useCallback(() => {
    if (!currentNode || !currentInstances) return;

    deletedModal({
      content: i18n.formatMessage(
        { id: "bigdata.workflow.delete.content" },
        { workflow: currentNode.name }
      ),
      onOk: () => {
        const hideMessage = message.loading(
          {
            content: i18n.formatMessage({
              id: "bigdata.workflow.delete.loading",
            }),
            key: "workflow",
          },
          0
        );

        deleteWorkflow
          .run(currentNode.id)
          .then((res) => {
            if (res?.code !== 0) {
              hideMessage();
              return;
            }
            if (selectNodeRef.current?.workflowId === currentNode?.id) {
              setSelectNode(undefined);
              onSetLocalData(null, LocalModuleType.dataAnalysisOpenNodeId);
              setUrlState({ nodeId: undefined });
            }
            getWorkflows.run({ iid: currentInstances! }).then((res) => {
              if (res?.code !== 0) {
                hideMessage();
                return;
              }
              setWorkflowList(res.data);
              message.success(
                {
                  content: i18n.formatMessage({
                    id: "bigdata.workflow.delete.success",
                  }),
                  key: "workflow",
                },
                3
              );
            });
          })
          .catch(() => hideMessage());
      },
    });
  }, [currentNode, currentInstances]);

  const createNodeModalCallback = useCallback((node: any) => {
    handleCloseNodeModal?.(node);
  }, []);

  const handleClickAddNode = useCallback(
    (
      primary: PrimaryEnums,
      secondary: SecondaryEnums,
      tertiary: TertiaryEnums
    ) => {
      if (!currentInstances) return;
      let extra: any = {
        iid: currentInstances,
        primary: primary,
        secondary: secondary,
        tertiary: tertiary,
        workflowId: currentNode?.workflowId || currentNode?.id,
      };
      extra.folderId =
        clickSource === OfflineRightMenuClickSourceEnums.folder
          ? currentNode?.id
          : currentNode?.folderId;
      setExtra(extra);
      showNodeModal(createNodeModalCallback);
    },
    [currentNode, currentInstances]
  );

  const handleClickUpdateNode = useCallback(() => {
    if (!currentInstances) return;
    setExtra({
      id: currentNode.id,
      iid: currentInstances,
      folderId: currentNode?.folderId,
      primary: currentNode?.primary,
      secondary: currentNode?.secondary,
      tertiary: currentNode?.tertiary,
    });
    setIsEditNode(true);
    setCurrentNode(currentNode);
    showNodeModal(handleCloseNodeModal);
  }, [currentNode, currentInstances]);

  const handleClickDeleteNode = useCallback(() => {
    if (!currentNode || !currentInstances) return;
    deletedModal({
      content: `确定删除节点${currentNode.name}吗？`,
      onOk: () => {
        const hideMessage = message.loading(
          {
            content: "删除中....",
            key: "node",
          },
          0
        );

        doDeletedNode
          .run(currentNode.id)
          .then((res) => {
            if (res?.code !== 0) {
              hideMessage();
              return;
            }
            handleCloseNodeModal?.();
            message.success(
              {
                content: "删除成功",
                key: "node",
              },
              3
            );
          })
          .catch(() => hideMessage());
      },
    });
  }, [currentNode, currentInstances]);

  const handleClickAddFolder = useCallback(
    (primary: PrimaryEnums, secondary: SecondaryEnums) => {
      if (!currentInstances) return;
      setExtra({
        iid: currentInstances,
        folderId: currentNode?.parentId,
        primary: primary,
        secondary: secondary,
        workflowId: currentNode?.id,
      });
      showFolderModal(handleCloseNodeModal);
    },
    [currentNode, currentInstances]
  );

  const handleClickUpdateFolder = useCallback(() => {
    if (!currentInstances) return;
    setExtra({
      id: currentNode.id,
      iid: currentInstances,
      folderId: currentNode?.parentId,
      primary: currentNode?.primary,
      secondary: currentNode?.secondary,
    });
    setIsEditNode(true);
    setCurrentNode(currentNode);
    showFolderModal(handleCloseNodeModal);
  }, [currentNode, currentInstances]);

  const handleClickDeleteFolder = useCallback(() => {
    if (!currentNode || !currentInstances) return;
    deletedModal({
      content: `确定删除文件夹${currentNode.name}吗？`,
      onOk: () => {
        const hideMessage = message.loading(
          {
            content: "删除中....",
            key: "folder",
          },
          0
        );

        doDeleteFolder
          .run(currentNode.id)
          .then((res) => {
            if (res?.code !== 0) {
              hideMessage();
              return;
            }
            handleCloseNodeModal?.();
            message.success(
              {
                content: "删除成功",
                key: "folder",
              },
              3
            );
          })
          .catch(() => hideMessage());
      },
    });
  }, [currentNode, currentInstances]);

  const workflowHeaderMenu: ItemType[] = [
    {
      label: i18n.formatMessage({ id: "bigdata.workflow.rightMenu.add" }),
      key: "add-workflow",
      icon: <AppstoreAddOutlined />,
      onClick: handleClickAddWorkflow,
    },
  ];

  const workflowMenu: ItemType[] = [
    // TODO: delete
    // {
    //   label: "看板",
    //   key: "workflow-board",
    //   icon: (
    //     <div style={{ marginRight: "8px", minWidth: "12px", fontSize: "12px" }}>
    //       <SVGIcon type={SVGTypeEnums.board} />
    //     </div>
    //   ),
    //   onClick: handleClickBoard,
    // },
    {
      label: i18n.formatMessage({ id: "bigdata.workflow.rightMenu.update" }),
      key: "update-workflow",
      icon: <EditOutlined />,
      onClick: handleClickUpdateWorkflow,
    },
    {
      label: (
        <span style={{ color: "hsl(0,68%,59%)" }}>
          {i18n.formatMessage({ id: "bigdata.workflow.rightMenu.delete" })}
        </span>
      ),
      key: "deleted-workflow",
      icon: <IconFont type={"icon-delete"} />,
      onClick: handleClickDeleteWorkflow,
    },
  ];
  const addNodeFromDataIntegration: ItemType[] = [
    {
      label: "新建节点",
      key: "add-node",
      children: [
        {
          label: "实时同步",
          key: "realTime-sync",
          onClick: () =>
            handleClickAddNode(
              PrimaryEnums.mining,
              SecondaryEnums.dataIntegration,
              TertiaryEnums.realtime
            ),
        },
        {
          label: "离线同步",
          key: "offline-sync",
          onClick: () =>
            handleClickAddNode(
              PrimaryEnums.mining,
              SecondaryEnums.dataIntegration,
              TertiaryEnums.offline
            ),
        },
      ],
    },
  ];

  const addFolder: ItemType[] = useMemo(() => {
    return [
      {
        label: "新建文件夹",
        key: "add-folder",
        onClick: () =>
          handleClickAddFolder(currentNode.primary, currentNode.secondary),
      },
    ];
  }, [currentNode]);

  const dataIntegrationMenu: ItemType[] = [
    ...addNodeFromDataIntegration,
    ...addFolder,
  ];

  const addNodeFromDevelopment: ItemType[] = [
    {
      label: "新建节点",
      key: "add-node",
      children: [
        {
          label: "MySQL",
          key: "MySQL",
          onClick: () =>
            handleClickAddNode(
              PrimaryEnums.mining,
              SecondaryEnums.dataMining,
              TertiaryEnums.mysql
            ),
        },
        {
          label: "ClickHouse",
          key: "ClickHouse",
          onClick: () =>
            handleClickAddNode(
              PrimaryEnums.mining,
              SecondaryEnums.dataMining,
              TertiaryEnums.clickhouse
            ),
        },
      ],
    },
  ];

  const dataDevelopmentMenu: ItemType[] = [
    ...addNodeFromDevelopment,
    ...addFolder,
  ];

  const nodeMenu: ItemType[] = [
    { label: "修改节点", key: "update-node", onClick: handleClickUpdateNode },
    { label: "删除节点", key: "delete-node", onClick: handleClickDeleteNode },
  ];

  const folderMenu: ItemType[] = [
    {
      label: "修改文件夹",
      key: "update-folder",
      onClick: () => handleClickUpdateFolder(),
    },
    {
      label: "删除文件夹",
      key: "delete-folder",
      onClick: () => handleClickDeleteFolder(),
    },
  ];

  const menuItems: ItemType[] = useMemo(() => {
    switch (clickSource) {
      case OfflineRightMenuClickSourceEnums.workflowHeader:
        return workflowHeaderMenu;
      case OfflineRightMenuClickSourceEnums.workflowItem:
        return workflowMenu;
      case OfflineRightMenuClickSourceEnums.dataIntegration:
        return dataIntegrationMenu;
      case OfflineRightMenuClickSourceEnums.dataDevelopment:
        return dataDevelopmentMenu;
      case OfflineRightMenuClickSourceEnums.node:
        return nodeMenu;
      case OfflineRightMenuClickSourceEnums.folder:
        let menu = lodash.cloneDeep(folderMenu);
        if (currentNode.secondary === SecondaryEnums.dataIntegration) {
          menu = [...addNodeFromDataIntegration, ...menu];
        }
        if (currentNode.secondary === SecondaryEnums.dataMining) {
          menu = [...addNodeFromDevelopment, ...menu];
        }
        return menu;
      default:
        return [];
    }
  }, [currentNode, clickSource]);

  return <Menu items={menuItems} />;
};
export default RightMenu;
